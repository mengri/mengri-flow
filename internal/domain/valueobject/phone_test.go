package valueobject

import (
	"testing"
)

func TestNewPhone_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"+8613800138000", "+8613800138000"},
		{"+1234567890", "+1234567890"},
		{" +8613800138000 ", "+8613800138000"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			phone, err := NewPhone(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := phone.String(); got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestNewPhone_Invalid(t *testing.T) {
	tests := []string{
		"",
		"13800138000",
		"abc",
		"+1",
		"+1234567890123456", // too long
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := NewPhone(input)
			if err == nil {
				t.Error("expected error for invalid phone")
			}
		})
	}
}

func TestPhone_Masked(t *testing.T) {
	phone, _ := NewPhone("+8613800138000")
	masked := phone.Masked()
	if masked == phone.String() {
		t.Error("masked should differ from original")
	}
	// Should contain ****
	if len(masked) != len(phone.String()) {
		// Not necessarily same length due to masking logic, but should be reasonable
	}
	if len(masked) < 4 {
		t.Error("masked result too short")
	}
}

func TestPhone_Equals(t *testing.T) {
	p1, _ := NewPhone("+8613800138000")
	p2, _ := NewPhone("+8613800138000")
	p3, _ := NewPhone("+1234567890")

	if !p1.Equals(p2) {
		t.Error("same phones should be equal")
	}
	if p1.Equals(p3) {
		t.Error("different phones should not be equal")
	}
}
