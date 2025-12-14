package handlers

import (
	"context"
	"io"
	"net/http"
)

type OpenAIHandler struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func (o *OpenAIHandler) ProviderName() string { return "openai" }

func (o *OpenAIHandler) SendRequest(ctx context.Context, payload []byte, apiKeyOverride string) ([]byte, error) {
	url := buildChatCompletionsURL(o.BaseURL)
	apiKey := o.APIKey
	if apiKeyOverride != "" {
		apiKey = apiKeyOverride
	}
	req, err := newJSONRequest(ctx, http.MethodPost, url, payload, apiKey)
	if err != nil {
		return nil, err
	}
	client := o.Client
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// Bubble up non-2xx responses as errors for the gateway to convert
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return body, &HTTPStatusError{StatusCode: res.StatusCode, Body: body}
	}
	return body, nil
}

type HTTPStatusError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPStatusError) Error() string {
	return http.StatusText(e.StatusCode)
}
