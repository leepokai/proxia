package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Provider    string
	ProviderURL string
	APIKey      string
	Port        string
	LogLevel    string
}

// LoadConfig attempts to load variables from a .env file (if present), then reads the environment.
func LoadConfig(dotenvPath string) (Config, error) {
	if dotenvPath != "" {
		if err := loadDotEnv(dotenvPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
	}
	return MustLoadFromEnv(), nil
}

func MustLoadFromEnv() Config {
	return Config{
		Provider:    strings.TrimSpace(os.Getenv("PROVIDER")),
		ProviderURL: strings.TrimSpace(os.Getenv("PROVIDER_URL")),
		APIKey:      strings.TrimSpace(os.Getenv("API_KEY")),
		Port:        strings.TrimSpace(os.Getenv("PORT")),
		LogLevel:    defaultString(strings.TrimSpace(os.Getenv("LOG_LEVEL")), "info"),
	}
}

func loadDotEnv(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		// Attempt relative to working directory's root if not found
		alt := filepath.Join(".", path)
		if alt != path {
			if _, err2 := os.Stat(alt); err2 == nil {
				path = alt
			} else {
				return err
			}
		} else {
			return err
		}
	} else if stat.IsDir() {
		// If a directory was passed, look for .env inside it
		path = filepath.Join(path, ".env")
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		// Support KEY=VALUE, allow quoted values
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		_ = os.Setenv(key, val)
	}
	return sc.Err()
}

func defaultString(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
