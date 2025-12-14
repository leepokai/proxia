package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// Derive32Key derives a 32-byte key from a string.
// If input is base64 and decodes to 32 bytes, it is used directly.
// Otherwise, sha256(input) is used.
func Derive32Key(s string) ([]byte, error) {
	if s == "" {
		return nil, errors.New("missing encryption key")
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) == 32 {
		return b, nil
	}
	sum := sha256.Sum256([]byte(s))
	return sum[:], nil
}

// DecryptString decrypts base64(nonce||ciphertext) produced by EncryptString.
func DecryptString(encBase64 string, secret string) (string, error) {
	key, err := Derive32Key(secret)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(encBase64)
	if err != nil {
		return "", errors.New("invalid encrypted payload")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted payload")
	}
	nonce := raw[:gcm.NonceSize()]
	ct := raw[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", errors.New("decrypt failed")
	}
	return string(pt), nil
}

// EncryptString encrypts plaintext using AES-256-GCM and returns base64(nonce||ciphertext).
// (Not currently used by the gateway, but shared for parity with the web app.)
func EncryptString(plaintext string, secret string) (string, error) {
	key, err := Derive32Key(secret)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	out := append(nonce, ct...)
	return base64.StdEncoding.EncodeToString(out), nil
}

