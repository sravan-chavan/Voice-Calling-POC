package config

import (
	"strings"
	"testing"
)

func TestConfigValidate_MissingRequired(t *testing.T) {
	cfg := &Config{}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	msg := err.Error()
	for _, key := range []string{"PLIVO_AUTH_ID", "PLIVO_AUTH_TOKEN", "PLIVO_FROM_NUMBER", "PLIVO_ANSWER_URL"} {
		if !strings.Contains(msg, key) {
			t.Errorf("expected missing %s in error, got %q", key, msg)
		}
	}
}

func TestConfigValidate_Valid(t *testing.T) {
	cfg := &Config{
		AuthID:       "MAXXXXXXXXXXXXXXXXXX",
		AuthToken:    "token",
		FromNumber:   "+14151234567",
		AnswerURL:    "https://example.com/answer.xml",
		AnswerMethod: "get",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AnswerMethod != "GET" {
		t.Fatalf("expected AnswerMethod normalized to GET, got %q", cfg.AnswerMethod)
	}
}

func TestConfigValidate_InvalidAnswerURL(t *testing.T) {
	cfg := &Config{
		AuthID:     "id",
		AuthToken:  "token",
		FromNumber: "+14151234567",
		AnswerURL:  "not-a-url",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid answer URL")
	}
}

func TestConfigValidate_InvalidAnswerMethod(t *testing.T) {
	cfg := &Config{
		AuthID:       "id",
		AuthToken:    "token",
		FromNumber:   "+14151234567",
		AnswerURL:    "https://example.com/answer.xml",
		AnswerMethod: "PUT",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid answer method")
	}
}

func TestRedacted_HidesToken(t *testing.T) {
	cfg := &Config{
		AuthID:     "MA1234567890",
		AuthToken:  "super-secret-token",
		FromNumber: "+14151234567",
		AnswerURL:  "https://example.com/answer.xml",
	}
	redacted := cfg.Redacted()
	if redacted["PLIVO_AUTH_TOKEN"] != "[REDACTED]" {
		t.Fatalf("token not redacted: %v", redacted["PLIVO_AUTH_TOKEN"])
	}
	if strings.Contains(redacted["PLIVO_AUTH_ID"], "1234567890") {
		t.Fatalf("auth id not masked: %v", redacted["PLIVO_AUTH_ID"])
	}
}
