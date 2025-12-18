package utils_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"goproject/utils"
)

func TestDerive32Key(t *testing.T) {
	t.Run("accepts base64 32 bytes", func(t *testing.T) {
		raw := bytes.Repeat([]byte{0x42}, 32)
		encoded := base64.StdEncoding.EncodeToString(raw)

		key, err := utils.Derive32Key(encoded)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Equal(key, raw) {
			t.Fatalf("expected decoded key to match input bytes")
		}
	})

	t.Run("falls back to sha256", func(t *testing.T) {
		input := "not-base64"
		key, err := utils.Derive32Key(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		sum := sha256.Sum256([]byte(input))
		if !bytes.Equal(key, sum[:]) {
			t.Fatalf("expected sha256 digest, got %x", key)
		}
	})

	t.Run("errors on empty input", func(t *testing.T) {
		_, err := utils.Derive32Key("")
		if err == nil {
			t.Fatalf("expected missing key error, got nil")
		}
		if err.Error() != "missing encryption key" {
			t.Fatalf("expected missing key error, got %v", err)
		}
	})
}

func TestEncryptDecryptString(t *testing.T) {
	secret := "super-secret"
	plaintext := "hello world"

	enc, err := utils.EncryptString(plaintext, secret)
	if err != nil {
		t.Fatalf("encrypt returned error: %v", err)
	}
	if enc == "" {
		t.Fatalf("expected encrypted output")
	}

	dec, err := utils.DecryptString(enc, secret)
	if err != nil {
		t.Fatalf("decrypt returned error: %v", err)
	}
	if dec != plaintext {
		t.Fatalf("expected decrypted text %q, got %q", plaintext, dec)
	}
}

func TestDecryptStringErrors(t *testing.T) {
	secret := "another-secret"

	t.Run("invalid base64 payload", func(t *testing.T) {
		_, err := utils.DecryptString("!!!not-base64!!!", secret)
		if err == nil || err.Error() != "invalid encrypted payload" {
			t.Fatalf("expected invalid payload error, got %v", err)
		}
	})

	t.Run("wrong secret fails to decrypt", func(t *testing.T) {
		enc, err := utils.EncryptString("data", "good-secret")
		if err != nil {
			t.Fatalf("encrypt returned error: %v", err)
		}
		_, err = utils.DecryptString(enc, secret)
		if err == nil || err.Error() != "decrypt failed" {
			t.Fatalf("expected decrypt failed error, got %v", err)
		}
	})
}
