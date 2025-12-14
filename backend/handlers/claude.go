package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ClaudeHandler forwards requests to Anthropic's Claude messages endpoint
// and returns an OpenAI-like normalized response.
type ClaudeHandler struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func (c *ClaudeHandler) ProviderName() string { return "claude" }

func (c *ClaudeHandler) SendRequest(ctx context.Context, payload []byte, apiKeyOverride string) ([]byte, error) {
	url := buildClaudeURL(c.BaseURL)
	apiKey := c.APIKey
	if apiKeyOverride != "" {
		apiKey = apiKeyOverride
	}
	converted, err := convertToAnthropicPayload(payload)
	if err != nil {
		return nil, err
	}
	req, err := newJSONRequest(ctx, http.MethodPost, url, converted, "")
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	// Anthropic requires version header
	req.Header.Set("anthropic-version", "2023-06-01")

	client := c.Client
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
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return body, &HTTPStatusError{StatusCode: res.StatusCode, Body: body}
	}

	// Convert Anthropic response to OpenAI-like shape for the gateway
	var aResp anthropicResponse
	if err := json.Unmarshal(body, &aResp); err != nil {
		return nil, err
	}
	text := aResp.JoinText()
	out := map[string]any{
		"id":       aResp.ID,
		"object":   "chat.completion",
		"created":  time.Now().Unix(),
		"model":    aResp.Model,
		"provider": "claude",
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": text,
				},
				"finish_reason": aResp.StopReason,
			},
		},
	}
	if aResp.Usage != nil {
		out["usage"] = map[string]any{
			"prompt_tokens":     aResp.Usage.InputTokens,
			"completion_tokens": aResp.Usage.OutputTokens,
			"total_tokens":      aResp.Usage.InputTokens + aResp.Usage.OutputTokens,
		}
	}
	return json.Marshal(out)
}

func buildClaudeURL(base string) string {
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/v1") {
		return base + "/messages"
	}
	return base + "/v1/messages"
}

// convertToAnthropicPayload transforms an OpenAI-style chat payload into
// Anthropic's /v1/messages format (including max_tokens requirement) and drops unknown fields (e.g., provider).
func convertToAnthropicPayload(raw []byte) ([]byte, error) {
	var in map[string]interface{}
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, err
	}
	// Remove fields Anthropic doesn't accept
	delete(in, "provider")

	model := toString(in["model"])
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}
	maxTokens := toInt(in["max_tokens"])
	if maxTokens <= 0 {
		maxTokens = 256
	}

	// Extract system message (if first is system)
	var systemStr string
	msgArr, _ := in["messages"].([]interface{})
	if len(msgArr) > 0 {
		if first, ok := msgArr[0].(map[string]interface{}); ok && strings.ToLower(toString(first["role"])) == "system" {
			systemStr = toStringContent(first["content"])
			msgArr = msgArr[1:]
		}
	}
	if systemStr == "" {
		systemStr = toStringContent(in["system"])
	}

	msgs := make([]anthropicMessage, 0, len(msgArr))
	for _, m := range msgArr {
		mm, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		role := strings.ToLower(toString(mm["role"]))
		if role == "" {
			role = "user"
		}
		text := toStringContent(mm["content"])
		if text == "" {
			continue
		}
		msgs = append(msgs, anthropicMessage{
			Role: role,
			Content: []anthropicContent{
				{Type: "text", Text: text},
			},
		})
	}

	out := anthropicPayload{
		Model:     model,
		MaxTokens: maxTokens,
		Messages:  msgs,
		System:    systemStr,
	}
	if t, ok := in["temperature"].(float64); ok {
		out.Temperature = &t
	}
	if t, ok := in["top_p"].(float64); ok {
		out.TopP = &t
	}
	return json.Marshal(out)
}

type anthropicPayload struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	Temperature *float64           `json:"temperature,omitempty"`
	TopP        *float64           `json:"top_p,omitempty"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicResponse struct {
	ID         string             `json:"id"`
	Model      string             `json:"model"`
	Content    []anthropicContent `json:"content"`
	StopReason string             `json:"stop_reason"`
	Usage      *anthropicUsage    `json:"usage"`
}

func (r anthropicResponse) JoinText() string {
	var parts []string
	for _, c := range r.Content {
		if c.Type == "text" && c.Text != "" {
			parts = append(parts, c.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func toStringContent(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []interface{}:
		var sb []string
		for _, item := range t {
			if s, ok := item.(string); ok {
				sb = append(sb, s)
			}
		}
		return strings.Join(sb, " ")
	case map[string]interface{}:
		if s, ok := t["text"].(string); ok {
			return s
		}
	}
	return ""
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	}
	return 0
}
