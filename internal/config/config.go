package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "fce-cli"
	keyringUser    = "api-key"
	configFileName = "config.json"
)

// Config holds persisted non-secret settings
type Config struct {
	FirstLogin bool   `json:"first_login"`
	Plan       string `json:"plan"`
	PlanLabel  string `json:"plan_label"`
}

// ── API key (stored in OS keyring) ───────────────────────────────────────────

func SaveAPIKey(key string) error {
	return keyring.Set(keyringService, keyringUser, key)
}

func LoadAPIKey() (string, error) {
	key, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		// Fallback: check env var
		if env := os.Getenv("FCE_API_KEY"); env != "" {
			return env, nil
		}
		return "", fmt.Errorf("not logged in — run: fce login")
	}
	return key, nil
}

func DeleteAPIKey() error {
	return keyring.Delete(keyringService, keyringUser)
}

func IsLoggedIn() bool {
	_, err := LoadAPIKey()
	return err == nil
}

// ── Config file (non-secret settings) ────────────────────────────────────────

func configDir() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "fce")
		}
	case "darwin":
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, "Library", "Application Support", "fce")
		}
	}
	// Linux / fallback
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "fce")
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "fce")
	}
	return filepath.Join(os.TempDir(), "fce")
}

func configPath() string {
	return filepath.Join(configDir(), configFileName)
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if os.IsNotExist(err) {
		return &Config{FirstLogin: true}, nil
	}
	if err != nil {
		return &Config{FirstLogin: true}, nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{FirstLogin: true}, nil
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}
