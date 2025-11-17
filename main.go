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
	if cfg.ProviderURL == "" && cfg.Provider == "openai" {
		cfg.ProviderURL = "https://api.openai.com/v1"
	}

	// Select provider based on config
	var provider handlers.AIProvider
	switch cfg.Provider {
	case "openai":
		provider = &handlers.OpenAIHandler{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.ProviderURL,
			Client:  &http.Client{Timeout: 60 * time.Second},
		}
	case "gemini":
		provider = &handlers.GeminiHandler{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.ProviderURL,
			Client:  &http.Client{Timeout: 60 * time.Second},
		}
	case "claude":
		provider = &handlers.ClaudeHandler{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.ProviderURL,
			Client:  &http.Client{Timeout: 60 * time.Second},
		}
	default:
		logger.Warnf("unknown provider %q; defaulting to openai", cfg.Provider)
		provider = &handlers.OpenAIHandler{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.ProviderURL,
			Client:  &http.Client{Timeout: 60 * time.Second},
		}
	}

	app := NewApp(cfg, provider, logger, startTime)
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
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("graceful shutdown failed: %v", err)
		_ = server.Close()
	}
}
