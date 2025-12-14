package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goproject/handlers"
	"goproject/ratelimit"
	"goproject/store"
	"goproject/utils"
)

var startTime = time.Now()

func main() {
	cfg, err := LoadConfig(".env")
	if err != nil {
		log.Printf("warning: failed to load .env: %v (continuing with environment)", err)
		// Continue; values may come from the real environment
		cfg = MustLoadFromEnv()
	}

	logger := utils.NewLogger(cfg.LogLevel)
	logger.Info("starting GoProject gateway...")

	// Defaults
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}

	// Build provider map for dynamic routing by model
	providers := make(map[string]handlers.AIProvider)

	openaiKey := cfg.OpenAIAPIKey
	if openaiKey == "" {
		openaiKey = cfg.APIKey
	}
	openaiURL := cfg.OpenAIURL
	if openaiURL == "" {
		openaiURL = "https://api.openai.com/v1"
	}
	providers["openai"] = &handlers.OpenAIHandler{
		APIKey:  openaiKey,
		BaseURL: openaiURL,
		Client:  &http.Client{Timeout: 60 * time.Second},
	}

	claudeKey := cfg.ClaudeAPIKey
	if claudeKey == "" {
		claudeKey = cfg.APIKey
	}
	claudeURL := cfg.ClaudeURL
	if claudeURL == "" {
		claudeURL = "https://api.anthropic.com"
	}
	providers["claude"] = &handlers.ClaudeHandler{
		APIKey:  claudeKey,
		BaseURL: claudeURL,
		Client:  &http.Client{Timeout: 60 * time.Second},
	}

	defaultProvider := cfg.Provider
	if _, ok := providers[defaultProvider]; !ok {
		if _, ok := providers["openai"]; ok {
			defaultProvider = "openai"
		} else {
			for name := range providers {
				defaultProvider = name
				break
			}
		}
	}

	// Initialize store (Postgres if available; otherwise in-memory for dev)
	var st store.Store
	if cfg.DatabaseURL != "" {
		logger.Infof("using postgres store: %s", cfg.DatabaseURL)
		s, err := store.NewPGStore(context.Background(), cfg.DatabaseURL)
		if err != nil {
			logger.Errorf("db init failed, falling back to memory store: %v", err)
			st = store.NewMemoryStore(cfg.DevGatewayKey)
		} else {
			st = s
		}
	} else {
		logger.Info("using in-memory store (set DATABASE_URL to enable Postgres-backed keys)")
		st = store.NewMemoryStore(cfg.DevGatewayKey)
	}
	lim := ratelimit.NewKeyedLimiter()

	app := NewApp(cfg, providers, defaultProvider, logger, startTime, st, lim)

	// Log a few keys from the store to confirm DB connectivity
	logSampleKeys(logger, st)

	addr := ":" + cfg.Port

	server := &http.Server{
		Addr:              addr,
		Handler:           app.mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       65 * time.Second,
		WriteTimeout:      65 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Infof("listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("server error: %v", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = st.Close(shutdownCtx)
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("graceful shutdown failed: %v", err)
		_ = server.Close()
	}
}

func logSampleKeys(logger *utils.Logger, st store.Store) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	keys, err := st.ListAPIKeys(ctx, 5)
	if err != nil {
		logger.Warnf("list api keys failed: %v", err)
		return
	}
	if len(keys) == 0 {
		logger.Infof("api keys: none found (check if DB has data or use /keys to create one)")
		return
	}
	for _, k := range keys {
		logger.Infof("api key: %s status=%s created=%s", k.KeyPrefix, k.Status, k.CreatedAt.Format(time.RFC3339))
	}
}
