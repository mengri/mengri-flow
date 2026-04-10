package valueobject

import (
	"testing"

	domainErr "mengri-flow/internal/domain/errors"
)

func TestValidatePasswordStrength_Valid(t *testing.T) {
	tests := []string{
		"Abcdef1!",
		"Password123!",
		"MyP@ssw0rd",
		"Str0ng!Pass",
		"aA1!aaaa",
	}
	for _, pwd := range tests {
		t.Run(pwd, func(t *testing.T) {
			if err := ValidatePasswordStrength(pwd); err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
		})
	}
}

func TestValidatePasswordStrength_Invalid(t *testing.T) {
	tests := []struct {
		name string
		pwd  string
	}{
		{"too short", "Aa1!"},
		{"no uppercase", "abcdef1!"},
		{"no lowercase", "ABCDEF1!"},
		{"no digit", "Abcdefgh!"},
		{"no special", "Abcdef12"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePasswordStrength(tt.pwd); err != domainErr.ErrInvalidPassword {
				t.Errorf("expected ErrInvalidPassword, got %v", err)
			}
		})
	}
}
