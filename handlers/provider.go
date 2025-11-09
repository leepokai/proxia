package handlers

import (
	"bytes"
	"context"
	"net/http"
	"strings"
)

type AIProvider interface {
	SendRequest(ctx context.Context, payload []byte) ([]byte, error)
	ProviderName() string
}

// buildChatCompletionsURL constructs the chat completions endpoint accounting
// for whether BaseURL already includes "/v1".
func buildChatCompletionsURL(base string) string {
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/v1") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}

func newJSONRequest(ctx context.Context, method, url string, payload []byte, apiKey string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	return req, nil
}
