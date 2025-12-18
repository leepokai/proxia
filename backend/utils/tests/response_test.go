package utils_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goproject/utils"
)

func TestNormalizeProviderResponseFillsDefaults(t *testing.T) {
	provider := "openai"
	raw := []byte(`{"choices":[]}`)

	out, err := utils.NormalizeProviderResponse(provider, raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("unmarshal normalized response: %v", err)
	}

	if m["provider"] != provider {
		t.Fatalf("expected provider %q, got %v", provider, m["provider"])
	}
	id, ok := m["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected generated id, got %v", m["id"])
	}
	if m["object"] != "chat.completion" {
		t.Fatalf("expected default object, got %v", m["object"])
	}
	created, ok := m["created"].(float64)
	if !ok || int64(created) <= 0 || int64(created) > time.Now().Unix() {
		t.Fatalf("expected created timestamp, got %v", m["created"])
	}
}

func TestNormalizeProviderResponseKeepsExisting(t *testing.T) {
	provider := "claude"
	raw := []byte(`{"id":"123","created":1700000000,"object":"custom","value":1}`)

	out, err := utils.NormalizeProviderResponse(provider, raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("unmarshal normalized response: %v", err)
	}
	if m["provider"] != provider {
		t.Fatalf("expected provider %q, got %v", provider, m["provider"])
	}
	if m["id"] != "123" {
		t.Fatalf("expected id to be preserved, got %v", m["id"])
	}
	if m["created"] != float64(1700000000) {
		t.Fatalf("expected created to be preserved, got %v", m["created"])
	}
	if m["object"] != "custom" {
		t.Fatalf("expected object to be preserved, got %v", m["object"])
	}
	if m["value"] != float64(1) {
		t.Fatalf("expected arbitrary fields preserved, got %v", m["value"])
	}
}

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	payload := struct {
		Foo string `json:"foo"`
	}{Foo: "bar"}

	utils.WriteJSON(rr, http.StatusCreated, payload)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", ct)
	}

	var decoded map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded["foo"] != "bar" {
		t.Fatalf("expected foo=bar, got %v", decoded["foo"])
	}
}

func TestWriteError(t *testing.T) {
	rr := httptest.NewRecorder()

	utils.WriteError(rr, http.StatusBadRequest, "invalid")

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", ct)
	}

	var body utils.ErrorBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Error.Code != http.StatusBadRequest || body.Error.Message != "invalid" {
		t.Fatalf("unexpected body: %+v", body)
	}
}
