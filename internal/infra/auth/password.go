package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 对明文密码进行哈希。
func HashPassword(plaintext string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword 验证明文密码与哈希是否匹配。
func VerifyPassword(plaintext, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	return err == nil
}
