package entity

import (
	"testing"
	"time"

	domainErr "mengri-flow/internal/domain/errors"
)

func TestNewAccount_Success(t *testing.T) {
	acc, err := NewAccount("test@example.com", "john", "John Doe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Status != AccountStatusPendingActivation {
		t.Errorf("expected status PENDING_ACTIVATION, got %s", acc.Status)
	}
	if acc.Role != "user" {
		t.Errorf("expected role user, got %s", acc.Role)
	}
	if acc.Username != "john" {
		t.Errorf("expected username john, got %s", acc.Username)
	}
	if acc.Email.String() != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", acc.Email.String())
	}
	if acc.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestNewAccount_Validation(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		username    string
		displayName string
		wantErr     error
	}{
		{"empty username", "a@b.com", "", "D", domainErr.ErrInvalidUsername},
		{"username too short", "a@b.com", "a", "D", domainErr.ErrInvalidUsername},
		{"username too long", "a@b.com", string(make([]byte, 51)), "D", domainErr.ErrInvalidUsername},
		{"empty displayName", "a@b.com", "john", "", domainErr.ErrInvalidDisplayName},
		{"invalid email", "not-an-email", "john", "D", domainErr.ErrInvalidEmail},
		{"empty email", "", "john", "D", domainErr.ErrInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAccount(tt.email, tt.username, tt.displayName)
			if err != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAccount_Activate(t *testing.T) {
	acc := &Account{Status: AccountStatusPendingActivation}

	err := acc.Activate("$hashed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Status != AccountStatusActive {
		t.Errorf("expected ACTIVE, got %s", acc.Status)
	}
	if acc.PasswordHash != "$hashed" {
		t.Errorf("password hash not set")
	}
	if acc.ActivatedAt == nil {
		t.Error("expected ActivatedAt to be set")
	}
}

func TestAccount_Activate_WrongStatus(t *testing.T) {
	acc := &Account{Status: AccountStatusActive}
	if err := acc.Activate("$hashed"); err != domainErr.ErrAlreadyActivated {
		t.Errorf("expected ErrAlreadyActivated, got %v", err)
	}
}

func TestAccount_Activate_EmptyPassword(t *testing.T) {
	acc := &Account{Status: AccountStatusPendingActivation}
	if err := acc.Activate(""); err != domainErr.ErrInvalidPassword {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestAccount_StateMachine(t *testing.T) {
	tests := []struct {
		name     string
		initial  AccountStatus
		action   func(*Account) error
		want     AccountStatus
		wantErr  error
	}{
		{"pending → active", AccountStatusPendingActivation, func(a *Account) error { return a.Activate("$h") }, AccountStatusActive, nil},
		{"active → locked", AccountStatusActive, func(a *Account) error { return a.Lock() }, AccountStatusLocked, nil},
		{"locked → active", AccountStatusLocked, func(a *Account) error { return a.Unlock() }, AccountStatusActive, nil},
		{"active → disabled", AccountStatusActive, func(a *Account) error { return a.Disable() }, AccountStatusDisabled, nil},
		{"locked → disabled", AccountStatusLocked, func(a *Account) error { return a.Disable() }, AccountStatusDisabled, nil},
		{"disabled → active", AccountStatusDisabled, func(a *Account) error { return a.Enable() }, AccountStatusActive, nil},
		// Invalid transitions
		{"active → active (lock)", AccountStatusActive, func(a *Account) error { return a.Unlock() }, AccountStatusActive, domainErr.ErrInvalidStatusTransition},
		{"pending → locked", AccountStatusPendingActivation, func(a *Account) error { return a.Lock() }, AccountStatusPendingActivation, domainErr.ErrInvalidStatusTransition},
		{"disabled → locked", AccountStatusDisabled, func(a *Account) error { return a.Lock() }, AccountStatusDisabled, domainErr.ErrInvalidStatusTransition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &Account{Status: tt.initial}
			err := tt.action(acc)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
			if acc.Status != tt.want {
				t.Errorf("expected status %s, got %s", tt.want, acc.Status)
			}
		})
	}
}

func TestAccount_CanLogin(t *testing.T) {
	tests := []struct {
		status AccountStatus
		want   bool
	}{
		{AccountStatusActive, true},
		{AccountStatusPendingActivation, false},
		{AccountStatusLocked, false},
		{AccountStatusDisabled, false},
	}
	for _, tt := range tests {
		acc := &Account{Status: tt.status}
		if got := acc.CanLogin(); got != tt.want {
			t.Errorf("CanLogin(%s) = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestAccount_ChangePassword(t *testing.T) {
	acc := &Account{Status: AccountStatusActive, PasswordHash: "old"}
	if err := acc.ChangePassword("new"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.PasswordHash != "new" {
		t.Errorf("password not updated")
	}

	// Non-active should fail
	acc.Status = AccountStatusLocked
	if err := acc.ChangePassword("newer"); err != domainErr.ErrInvalidStatusTransition {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}

	// Empty password should fail
	acc.Status = AccountStatusActive
	if err := acc.ChangePassword(""); err != domainErr.ErrInvalidPassword {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestAccount_IsAdmin(t *testing.T) {
	admin := &Account{Role: "admin"}
	user := &Account{Role: "user"}
	if !admin.IsAdmin() {
		t.Error("admin should be admin")
	}
	if user.IsAdmin() {
		t.Error("user should not be admin")
	}
}

func TestAccount_LowercaseEmail(t *testing.T) {
	acc, err := NewAccount("John@Example.COM", "john", "John")
	if err != nil {
		t.Fatal(err)
	}
	if acc.Email.String() != "john@example.com" {
		t.Errorf("expected lowercase email, got %s", acc.Email.String())
	}
}

func TestAccount_UpdateTimestamps(t *testing.T) {
	acc := &Account{Status: AccountStatusActive, UpdatedAt: time.Now().Add(-time.Hour)}
	_ = acc.Lock()
	if acc.UpdatedAt.Before(time.Now().Add(-time.Minute)) {
		t.Error("UpdatedAt should be refreshed")
	}
}
