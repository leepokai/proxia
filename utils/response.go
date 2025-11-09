package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// NormalizeProviderResponse ensures a uniform response shape by injecting
// "provider", and filling missing "id" and "created" when possible.
func NormalizeProviderResponse(provider string, raw []byte) ([]byte, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	m["provider"] = provider
	if _, ok := m["id"]; !ok {
		m["id"] = "gateway-" + shortID()
	}
	if _, ok := m["created"]; !ok {
		m["created"] = time.Now().Unix()
	}
	if _, ok := m["object"]; !ok {
		// Reasonable default for non-streaming chat
		m["object"] = "chat.completion"
	}
	out, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func shortID() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
