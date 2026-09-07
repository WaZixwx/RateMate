// Package config handles loading and persisting the RateMate configuration
// to ~/.ratemate/config.json.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WaZixwx/RateMate/internal/i18n"
	"github.com/WaZixwx/RateMate/internal/limiter"
)

// File is the on-disk representation. It is a superset of the limiter config
// plus proxy-level settings (port, target filtering, etc).
type File struct {
	Mode          string `json:"mode"`            // "auto" | "manual"
	Window        string `json:"window"`          // "second" | "minute" | "hour"
	Limit         int    `json:"limit"`           // requests per window (auto)
	ManualDelayMs int    `json:"manual_delay_ms"` // ms (manual)
	Port          int    `json:"port"`            // proxy listen port
	AutoStart     bool   `json:"autostart"`       // start proxy on launch
	LogBodyBytes  bool   `json:"log_body_bytes"`  // (reserved) log body sizes
	Language      string `json:"language"`        // UI language code (e.g. "en","zh")
}

// Default returns a sensible starting configuration.
func Default() File {
	return File{
		Mode:          "auto",
		Window:        "minute",
		Limit:         50,
		ManualDelayMs: 1000,
		Port:          8080,
		AutoStart:     false,
		Language:      string(i18n.Default),
	}
}

// ToLimiterConfig converts the on-disk file into a limiter.Config.
func (f File) ToLimiterConfig() limiter.Config {
	cfg := limiter.Config{
		Limit:         f.Limit,
		ManualDelayMs: f.ManualDelayMs,
	}
	switch f.Mode {
	case "manual":
		cfg.Mode = limiter.ModeManual
	default:
		cfg.Mode = limiter.ModeAuto
	}
	switch f.Window {
	case "second":
		cfg.Window = limiter.WindowSecond
	case "hour":
		cfg.Window = limiter.WindowHour
	default:
		cfg.Window = limiter.WindowMinute
	}
	return cfg
}

// FromLimiterConfig reflects a limiter.Config back into the on-disk file.
func (f *File) FromLimiterConfig(c limiter.Config) {
	f.Limit = c.Limit
	f.ManualDelayMs = c.ManualDelayMs
	switch c.Mode {
	case limiter.ModeManual:
		f.Mode = "manual"
	default:
		f.Mode = "auto"
	}
	switch c.Window {
	case limiter.WindowSecond:
		f.Window = "second"
	case limiter.WindowHour:
		f.Window = "hour"
	default:
		f.Window = "minute"
	}
}

// configPath returns the absolute path to the config file, creating the
// directory if necessary.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".ratemate")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config file, falling back to Default() when it is missing or
// malformed.
func Load() File {
	p, err := configPath()
	if err != nil {
		return Default()
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return Default()
	}
	f := Default()
	if err := json.Unmarshal(data, &f); err != nil {
		return Default()
	}
	// Normalise / clamp.
	if f.Port <= 0 || f.Port > 65535 {
		f.Port = 8080
	}
	if f.Limit < 1 {
		f.Limit = 1
	}
	if f.ManualDelayMs < 0 {
		f.ManualDelayMs = 0
	}
	f.Language = string(i18n.Normalise(i18n.Lang(f.Language)))
	return f
}

// Save persists the config file, returning an error if writing fails.
func Save(f File) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
