package entity

import (
	"testing"

	domainErr "mengri-flow/internal/domain/errors"
)

func TestNewIdentity_Success(t *testing.T) {
	id, err := NewIdentity("acc-1", LoginTypePassword, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.AccountID != "acc-1" {
		t.Errorf("expected accountID acc-1, got %s", id.AccountID)
	}
	if id.LoginType != LoginTypePassword {
		t.Errorf("expected loginType password, got %s", id.LoginType)
	}
	if id.ExternalID != "user@example.com" {
		t.Errorf("expected externalID user@example.com, got %s", id.ExternalID)
	}
}

func TestNewIdentity_EmptyAccountID(t *testing.T) {
	_, err := NewIdentity("", LoginTypePassword, "user@example.com")
	if err != domainErr.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestNewIdentity_EmptyExternalID(t *testing.T) {
	_, err := NewIdentity("acc-1", LoginTypePassword, "")
	if err != domainErr.ErrIdentityNotBound {
		t.Errorf("expected ErrIdentityNotBound, got %v", err)
	}
}

func TestCanUnbind(t *testing.T) {
	tests := []struct {
		name          string
		activeCount   int
		canUnbind     bool
	}{
		{"single identity", 1, false},
		{"multiple identities", 2, true},
		{"zero identities", 0, false},
		{"many identities", 5, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanUnbind(tt.activeCount); got != tt.canUnbind {
				t.Errorf("CanUnbind(%d) = %v, want %v", tt.activeCount, got, tt.canUnbind)
			}
		})
	}
}

func TestLoginType_Constants(t *testing.T) {
	types := []LoginType{LoginTypePassword, LoginTypeSMS, LoginTypeWechatQR, LoginTypeLarkQR, LoginTypeGithub}
	for _, lt := range types {
		if string(lt) == "" {
			t.Errorf("login type should not be empty")
		}
	}
}
