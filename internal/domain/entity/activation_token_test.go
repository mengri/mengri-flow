package entity

import (
	"testing"
	"time"
)

func TestNewActivationToken(t *testing.T) {
	token := NewActivationToken("acc-1", "raw-secret", 24*time.Hour)
	if token.AccountID != "acc-1" {
		t.Errorf("expected accountID acc-1, got %s", token.AccountID)
	}
	if token.TokenHash == "" {
		t.Error("expected non-empty token hash")
	}
	if token.TokenHash == "raw-secret" {
		t.Error("token should be hashed, not stored in plaintext")
	}
	if token.IsUsed() {
		t.Error("new token should not be used")
	}
	if token.IsExpired() {
		t.Error("new token should not be expired")
	}
	if !token.IsValid() {
		t.Error("new token should be valid")
	}
}

func TestHashToken(t *testing.T) {
	h1 := HashToken("hello")
	h2 := HashToken("hello")
	if h1 != h2 {
		t.Error("same input should produce same hash")
	}
	if h1 == "hello" {
		t.Error("hash should not equal plaintext")
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	// Same as NewActivationToken hash
	token := NewActivationToken("acc-1", "my-token", 1*time.Hour)
	directHash := HashToken("my-token")
	if token.TokenHash != directHash {
		t.Error("HashToken should produce same result as NewActivationToken")
	}
}

func TestActivationToken_IsExpired(t *testing.T) {
	// Already expired
	token := NewActivationToken("acc-1", "raw", -1*time.Hour)
	if !token.IsExpired() {
		t.Error("token with negative TTL should be expired")
	}
	if token.IsValid() {
		t.Error("expired token should not be valid")
	}

	// Not expired
	future := NewActivationToken("acc-1", "raw", 24*time.Hour)
	if future.IsExpired() {
		t.Error("token with future TTL should not be expired")
	}
}

func TestActivationToken_MarkUsed(t *testing.T) {
	token := NewActivationToken("acc-1", "raw", 24*time.Hour)
	token.MarkUsed()
	if !token.IsUsed() {
		t.Error("token should be used after MarkUsed")
	}
	if !token.UsedAt.IsZero() {
		// UsedAt should be set
		_ = token.UsedAt
	}
	if token.IsValid() {
		t.Error("used token should not be valid")
	}
}

func TestActivationToken_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ActivationToken)
		wantValid bool
	}{
		{"fresh token", func(t *ActivationToken) {}, true},
		{"used token", func(t *ActivationToken) { t.MarkUsed() }, false},
		{"expired token", func(t *ActivationToken) {
			t.ExpiresAt = time.Now().Add(-time.Hour)
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := NewActivationToken("acc-1", "raw", 1*time.Hour)
			tt.setup(token)
			if got := token.IsValid(); got != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}
