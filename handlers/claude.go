package handlers

import (
	"context"
	"errors"
	"net/http"
)

type ClaudeHandler struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func (c *ClaudeHandler) ProviderName() string { return "claude" }

func (c *ClaudeHandler) SendRequest(ctx context.Context, payload []byte) ([]byte, error) {
	return nil, errors.New("claude handler not implemented yet")
}
