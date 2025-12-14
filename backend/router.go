package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"goproject/auth"
	"goproject/handlers"
	"goproject/ratelimit"
	"goproject/store"
	"goproject/utils"
)

type App struct {
	mux       *http.ServeMux
	cfg       Config
	providers map[string]handlers.AIProvider
	defaultPr string
	logger    *utils.Logger
	startTime time.Time
	store     store.Store
	limiter   *ratelimit.KeyedLimiter
}

func NewApp(cfg Config, providers map[string]handlers.AIProvider, defaultProvider string, logger *utils.Logger, startTime time.Time, st store.Store, lim *ratelimit.KeyedLimiter) *App {
	app := &App{
		mux:       http.NewServeMux(),
		cfg:       cfg,
		providers: providers,
		defaultPr: defaultProvider,
		logger:    logger,
		startTime: startTime,
		store:     st,
		limiter:   lim,
	}
	app.routes()
	return app
}

func (a *App) routes() {
	// Protected routes
	secured := auth.Middleware(a.store, a.limiter, a.logger)
	chatHandler := corsMiddleware(loggingMiddleware(a.logger, secured(http.HandlerFunc(a.handleChat))))
	a.mux.Handle("POST /v1/chat", chatHandler)
	a.mux.Handle("OPTIONS /v1/chat", chatHandler) // preflight

	common := func(h http.HandlerFunc) http.Handler {
		return corsMiddleware(loggingMiddleware(a.logger, h))
	}
	a.mux.Handle("GET /v1/config", common(a.handleConfig))
	a.mux.Handle("GET /v1/health", common(a.handleHealth))
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(a.startTime).Round(time.Second).String()
	resp := map[string]string{
		"status": "ok",
		"uptime": uptime,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (a *App) handleConfig(w http.ResponseWriter, r *http.Request) {
	type cfgResp struct {
		DefaultProvider    string   `json:"default_provider"`
		AvailableProviders []string `json:"available_providers"`
		ProviderURL        string   `json:"provider_url"`
		Port               string   `json:"port"`
		LogLevel           string   `json:"log_level"`
	}
	var names []string
	for name := range a.providers {
		names = append(names, name)
	}
	resp := cfgResp{
		DefaultProvider:    a.defaultPr,
		AvailableProviders: names,
		ProviderURL:        a.cfg.ProviderURL,
		Port:               a.cfg.Port,
		LogLevel:           a.cfg.LogLevel,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (a *App) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	defer r.Body.Close()

	var payload json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	// detect provider/model for routing
	var meta struct {
		Provider    string `json:"provider"`
		Model       string `json:"model"`
		BaseURL     string `json:"base_url"`
		ProviderURL string `json:"provider_url"` // alias for compatibility
	}
	_ = json.Unmarshal(payload, &meta)
	provName, err := a.selectProviderWithHint(meta.Provider, meta.Model)
	if err != nil || provName == "" {
		utils.WriteError(w, http.StatusBadRequest, "no provider available for this request")
		return
	}
	provider := a.providers[provName]
	baseURL := strings.TrimSpace(firstNonEmpty(meta.BaseURL, meta.ProviderURL))
	if baseURL != "" {
		provider = overrideProviderBaseURL(provider, baseURL)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Upstream key selection priority:
	// 1) Raw upstream key supplied as Bearer when not a gateway key (pass-through).
	// 2) BYOK stored per-user (when gateway key + encryption enabled).
	// 3) Server-level configured keys.
	var upstreamKey string

	// 1) Direct upstream key (non-gateway bearer).
	if uk := auth.UpstreamKeyFromContext(r.Context()); uk != "" {
		upstreamKey = uk
	}

	// 2) BYOK for authenticated gateway users.
	if upstreamKey == "" {
		if k := auth.APIKeyFromContext(r.Context()); k != nil && a.cfg.EncryptionKey != "" {
			if enc, err := a.store.GetProviderCredential(ctx, k.UserID, provider.ProviderName(), "default"); err == nil && enc != "" {
				if dec, err := utils.DecryptString(enc, a.cfg.EncryptionKey); err == nil {
					upstreamKey = dec
				}
			}
		}
	}

	// 3) Fallback to server-level upstream keys
	if upstreamKey == "" {
		switch provider.ProviderName() {
		case "openai":
			upstreamKey = firstNonEmpty(a.cfg.OpenAIAPIKey, a.cfg.APIKey)
		case "claude":
			upstreamKey = firstNonEmpty(a.cfg.ClaudeAPIKey, a.cfg.APIKey)
		default:
			upstreamKey = a.cfg.APIKey
		}
	}
	if upstreamKey == "" {
		utils.WriteError(w, http.StatusUnauthorized, "missing upstream API key")
		return
	}

	// Remove gateway-only fields before forwarding upstream.
	cleanPayload := payload
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err == nil {
		delete(m, "provider")
		delete(m, "base_url")
		delete(m, "provider_url")
		if b, err := json.Marshal(m); err == nil {
			cleanPayload = b
		}
	}

	respBytes, err := provider.SendRequest(ctx, cleanPayload, upstreamKey)
	if err != nil {
		// Return clearer errors depending on the failure
		if nie, ok := err.(*handlers.NotImplementedError); ok {
			a.logger.Warnf("provider not implemented: %s", nie.Error())
			utils.WriteError(w, http.StatusNotImplemented, nie.Error())
			return
		}
		if hse, ok := err.(*handlers.HTTPStatusError); ok {
			// Attempt to extract a readable message from provider body
			msg := extractProviderMessage(hse.Body)
			if msg == "" {
				msg = fmt.Sprintf("upstream status %d", hse.StatusCode)
			}
			a.logger.Warnf("upstream provider error (%d): %s", hse.StatusCode, msg)
			utils.WriteError(w, hse.StatusCode, "provider error: "+msg)
			return
		}
		a.logger.Errorf("provider error: %v", err)
		utils.WriteError(w, http.StatusBadGateway, "failed to contact provider")
		return
	}

	normalized, err := utils.NormalizeProviderResponse(provider.ProviderName(), respBytes)
	if err != nil {
		a.logger.Errorf("normalize error: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "failed to normalize response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(normalized)
}

// extractProviderMessage tries to derive a concise error message from an upstream response body.
func extractProviderMessage(body []byte) string {
	var m map[string]any
	if err := json.Unmarshal(body, &m); err == nil {
		// Common shapes: {"error": {"message": "..."}}, {"message": "..."}
		if errField, ok := m["error"]; ok {
			switch e := errField.(type) {
			case map[string]any:
				if msg, ok := e["message"].(string); ok && msg != "" {
					return msg
				}
				// anthropic-like {"error":{"type":"...","message":"..."}}
				if msg, ok := e["error"].(string); ok && msg != "" {
					return msg
				}
			case string:
				if e != "" {
					return e
				}
			}
		}
		if msg, ok := m["message"].(string); ok && msg != "" {
			return msg
		}
	}
	// fallback to raw body as string (may be JSON or text)
	return string(body)
}

// selectProvider picks provider based on model name; fallbacks to default.
func (a *App) selectProvider(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))
	// Route claude-like models
	if strings.HasPrefix(m, "claude") || strings.Contains(m, "anthropic") {
		if _, ok := a.providers["claude"]; ok {
			return "claude"
		}
	}
	// Route openai-like models
	if strings.HasPrefix(m, "gpt") || strings.Contains(m, "openai") || strings.HasPrefix(m, "o1") {
		if _, ok := a.providers["openai"]; ok {
			return "openai"
		}
	}
	// Default provider if available
	if _, ok := a.providers[a.defaultPr]; ok {
		return a.defaultPr
	}
	for name := range a.providers {
		return name
	}
	return ""
}

func (a *App) selectProviderWithHint(hint, model string) (string, error) {
	h := strings.ToLower(strings.TrimSpace(hint))
	if h != "" {
		if _, ok := a.providers[h]; ok {
			return h, nil
		}
		return "", fmt.Errorf("provider %q not available", h)
	}
	return a.selectProvider(model), nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// overrideProviderBaseURL returns a shallow copy of the provider with BaseURL replaced when supported.
func overrideProviderBaseURL(p handlers.AIProvider, baseURL string) handlers.AIProvider {
	switch t := p.(type) {
	case *handlers.OpenAIHandler:
		cp := *t
		cp.BaseURL = baseURL
		return &cp
	case *handlers.ClaudeHandler:
		cp := *t
		cp.BaseURL = baseURL
		return &cp
	default:
		// Unknown provider types just use the original (no override support).
		return p
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(logger *utils.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		duration := time.Since(start).Milliseconds()
		logger.Infof("%d %s %s %dms", sw.status, r.Method, r.URL.Path, duration)
	})
}

// corsMiddleware adds permissive CORS headers for browser clients (e.g., web playground).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
