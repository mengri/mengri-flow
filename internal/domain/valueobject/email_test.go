package valueobject

import (
	"testing"

	domainErr "mengri-flow/internal/domain/errors"
)

func TestNewEmail_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user@example.com", "user@example.com"},
		{"User@Example.COM", "user@example.com"},
		{"  user@example.com  ", "user@example.com"},
		{"test+tag@domain.co.uk", "test+tag@domain.co.uk"},
		{"a@b.co", "a@b.co"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			email, err := NewEmail(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := email.String(); got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestNewEmail_Invalid(t *testing.T) {
	tests := []string{
		"",
		"not-an-email",
		"@domain.com",
		"user@",
		"user@.com",
		"spaces in@email.com",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := NewEmail(input)
			if err != domainErr.ErrInvalidEmail {
				t.Errorf("expected ErrInvalidEmail, got %v", err)
			}
		})
	}
}

func TestEmail_Equals(t *testing.T) {
	e1, _ := NewEmail("user@example.com")
	e2, _ := NewEmail("user@example.com")
	e3, _ := NewEmail("other@example.com")

	if !e1.Equals(e2) {
		t.Error("same emails should be equal")
	}
	if e1.Equals(e3) {
		t.Error("different emails should not be equal")
	}
}
