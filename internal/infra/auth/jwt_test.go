package auth

import (
	"testing"
	"time"

	"mengri-flow/internal/infra/config"
)

func newTestJWTManager() *JWTManagerImpl {
	return &JWTManagerImpl{
		secret:             []byte("test-secret-key-for-testing"),
		accessTokenExpiry:  time.Hour,
		refreshTokenExpiry: 24 * time.Hour,
	}
}

func TestGenerateAndParseAccessToken(t *testing.T) {
	mgr := newTestJWTManager()

	token, err := mgr.GenerateAccessToken("acc-1", "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := mgr.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.AccountID != "acc-1" {
		t.Errorf("expected accountID acc-1, got %s", claims.AccountID)
	}
	if claims.Role != "user" {
		t.Errorf("expected role user, got %s", claims.Role)
	}
}

func TestGenerateAndParseRefreshToken(t *testing.T) {
	mgr := newTestJWTManager()

	token, err := mgr.GenerateRefreshToken("acc-2", "admin")
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	claims, err := mgr.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.AccountID != "acc-2" {
		t.Errorf("expected accountID acc-2, got %s", claims.AccountID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	mgr := newTestJWTManager()

	_, err := mgr.ParseToken("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	mgr1 := &JWTManagerImpl{
		secret:             []byte("secret-1"),
		accessTokenExpiry:  time.Hour,
		refreshTokenExpiry: 24 * time.Hour,
	}
	mgr2 := &JWTManagerImpl{
		secret:             []byte("secret-2"),
		accessTokenExpiry:  time.Hour,
		refreshTokenExpiry: 24 * time.Hour,
	}

	token, _ := mgr1.GenerateAccessToken("acc-1", "user")
	_, err := mgr2.ParseToken(token)
	if err == nil {
		t.Error("expected error when parsing token signed with different secret")
	}
}

func TestAccessTokenExpiry(t *testing.T) {
	mgr := newTestJWTManager()
	if mgr.AccessTokenExpiry() != int(time.Hour.Seconds()) {
		t.Errorf("expected %d, got %d", int(time.Hour.Seconds()), mgr.AccessTokenExpiry())
	}
}

func TestRefreshTokenExpiry(t *testing.T) {
	mgr := newTestJWTManager()
	if mgr.RefreshTokenExpiry() != 24*time.Hour {
		t.Errorf("expected 24h, got %v", mgr.RefreshTokenExpiry())
	}
}

func TestGenerateJWTManager_FromConfig(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:             "test-secret",
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 86400,
	}
	// Just verify it doesn't panic
	GenerateJWTManager(cfg)
}
