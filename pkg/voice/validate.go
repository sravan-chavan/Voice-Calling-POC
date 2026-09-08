package voice

import (
	"fmt"
	"regexp"
	"strings"
)

// E.164: optional +, then country code starting 1-9, then up to 14 more digits.
var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

// ValidateInitiateCallRequest validates a call request before contacting a provider.
func ValidateInitiateCallRequest(req InitiateCallRequest) error {
	to := strings.TrimSpace(req.To)
	if to == "" {
		return NewError(ErrInvalidInput, "destination number (to) is required", nil)
	}
	if !e164Pattern.MatchString(to) {
		return NewError(
			ErrInvalidInput,
			fmt.Sprintf("destination number must be E.164 format (e.g. +14155552671), got %q", to),
			nil,
		)
	}
	return nil
}

// NormalizePhone trims whitespace from a phone number string.
func NormalizePhone(number string) string {
	return strings.TrimSpace(number)
}
