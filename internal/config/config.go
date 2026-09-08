package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds runtime configuration for the voice calling module and POC server.
type Config struct {
	AuthID       string
	AuthToken    string
	FromNumber   string
	AnswerURL    string
	AnswerMethod string
	// CallTo is optional; used by the standalone CLI when -to is not provided.
	CallTo string
	// HTTPAddr when set enables HTTP server mode (e.g. ":8080").
	HTTPAddr string
}

// Load reads configuration from the process environment.
// If a local .env file exists it is loaded first (non-fatal if missing),
// so production can rely solely on injected environment variables.
func Load() (*Config, error) {
	_ = godotenv.Load() // ignore error: .env is optional

	cfg := &Config{
		AuthID:       strings.TrimSpace(os.Getenv("PLIVO_AUTH_ID")),
		AuthToken:    strings.TrimSpace(os.Getenv("PLIVO_AUTH_TOKEN")),
		FromNumber:   strings.TrimSpace(os.Getenv("PLIVO_FROM_NUMBER")),
		AnswerURL:    strings.TrimSpace(os.Getenv("PLIVO_ANSWER_URL")),
		AnswerMethod: strings.TrimSpace(os.Getenv("PLIVO_ANSWER_METHOD")),
		CallTo:       strings.TrimSpace(os.Getenv("CALL_TO")),
		HTTPAddr:     strings.TrimSpace(os.Getenv("HTTP_ADDR")),
	}

	if cfg.AnswerMethod == "" {
		cfg.AnswerMethod = "GET"
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate ensures required configuration is present and usable.
func (c *Config) Validate() error {
	var missing []string

	if c.AuthID == "" {
		missing = append(missing, "PLIVO_AUTH_ID")
	}
	if c.AuthToken == "" {
		missing = append(missing, "PLIVO_AUTH_TOKEN")
	}
	if c.FromNumber == "" {
		missing = append(missing, "PLIVO_FROM_NUMBER")
	}
	if c.AnswerURL == "" {
		missing = append(missing, "PLIVO_ANSWER_URL")
	}

	if len(missing) > 0 {
		return fmt.Errorf("configuration error: missing required environment variables: %s", strings.Join(missing, ", "))
	}

	method := strings.ToUpper(c.AnswerMethod)
	if method != "GET" && method != "POST" {
		return fmt.Errorf("configuration error: PLIVO_ANSWER_METHOD must be GET or POST, got %q", c.AnswerMethod)
	}
	c.AnswerMethod = method

	if !strings.HasPrefix(c.AnswerURL, "http://") && !strings.HasPrefix(c.AnswerURL, "https://") {
		return fmt.Errorf("configuration error: PLIVO_ANSWER_URL must be an absolute http(s) URL")
	}

	return nil
}

// Redacted returns a copy safe for logging (secrets removed).
func (c *Config) Redacted() map[string]string {
	return map[string]string{
		"PLIVO_AUTH_ID":       mask(c.AuthID),
		"PLIVO_AUTH_TOKEN":    "[REDACTED]",
		"PLIVO_FROM_NUMBER":   c.FromNumber,
		"PLIVO_ANSWER_URL":    c.AnswerURL,
		"PLIVO_ANSWER_METHOD": c.AnswerMethod,
		"HTTP_ADDR":           c.HTTPAddr,
	}
}

func mask(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}
