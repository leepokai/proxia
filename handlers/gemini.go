package handlers

import (
	"context"
	"errors"
	"net/http"
)

type GeminiHandler struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func (g *GeminiHandler) ProviderName() string { return "gemini" }

func (g *GeminiHandler) SendRequest(ctx context.Context, payload []byte) ([]byte, error) {
	return nil, errors.New("gemini handler not implemented yet")
}
