package handlers

import (
	"context"
	"net/http"
)

type GeminiHandler struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func (g *GeminiHandler) ProviderName() string { return "gemini" }

func (g *GeminiHandler) SendRequest(ctx context.Context, payload []byte, apiKeyOverride string) ([]byte, error) {
	return nil, &NotImplementedError{Provider: "gemini"}
}
