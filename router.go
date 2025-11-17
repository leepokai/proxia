package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"goproject/handlers"
	"goproject/utils"
)

type App struct {
	mux       *http.ServeMux
	cfg       Config
	provider  handlers.AIProvider
	logger    *utils.Logger
	startTime time.Time
}

func NewApp(cfg Config, provider handlers.AIProvider, logger *utils.Logger, startTime time.Time) *App {
	app := &App{
		mux:       http.NewServeMux(),
		cfg:       cfg,
		provider:  provider,
		logger:    logger,
		startTime: startTime,
	}
	app.routes()
	return app
}

func (a *App) routes() {
	a.mux.HandleFunc("POST /v1/chat", a.handleChat)
	a.mux.HandleFunc("GET /v1/config", a.handleConfig)
	a.mux.HandleFunc("GET /v1/health", a.handleHealth)
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
		Provider    string `json:"provider"`
		ProviderURL string `json:"provider_url"`
		Port        string `json:"port"`
		LogLevel    string `json:"log_level"`
	}
	resp := cfgResp{
		Provider:    a.provider.ProviderName(),
		ProviderURL: a.cfg.ProviderURL,
		Port:        a.cfg.Port,
		LogLevel:    a.cfg.LogLevel,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (a *App) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if a.cfg.APIKey == "" {
		utils.WriteError(w, http.StatusUnauthorized, "missing API key")
		return
	}

	defer r.Body.Close()

	var payload json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	respBytes, err := a.provider.SendRequest(ctx, payload)
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

	normalized, err := utils.NormalizeProviderResponse(a.provider.ProviderName(), respBytes)
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
