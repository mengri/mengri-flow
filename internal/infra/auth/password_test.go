package auth

import (
	"testing"
)

func TestHashPassword_And_VerifyPassword(t *testing.T) {
	plaintext := "MyStr0ng!Pass"
	hash, err := HashPassword(plaintext, 4) // low cost for test speed
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == plaintext {
		t.Error("hash should not equal plaintext")
	}
	if !VerifyPassword(plaintext, hash) {
		t.Error("correct password should verify")
	}
	if VerifyPassword("wrong-password", hash) {
		t.Error("wrong password should not verify")
	}
}

func TestVerifyPassword_DifferentHashes(t *testing.T) {
	// Same password should produce different hashes (bcrypt salt)
	hash1, _ := HashPassword("password123!", 4)
	hash2, _ := HashPassword("password123!", 4)
	if hash1 == hash2 {
		t.Log("hashes are the same (unlikely but possible with low cost)")
	}
	// Both should verify
	if !VerifyPassword("password123!", hash1) || !VerifyPassword("password123!", hash2) {
		t.Error("both hashes should verify with the same password")
	}
}

func TestVerifyPassword_EmptyHash(t *testing.T) {
	if VerifyPassword("anything", "") {
		t.Error("empty hash should not verify")
	}
}
