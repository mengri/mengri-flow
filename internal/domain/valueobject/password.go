package valueobject

import (
	"mengri-flow/internal/domain/errors"
	"regexp"
	"unicode"
)

// PasswordStrength 密码强度校验规则。
var passwordSpecialChars = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`)

// ValidatePasswordStrength 验证密码是否满足强度要求：
// - 至少 8 个字符
// - 至少 1 个大写字母
// - 至少 1 个小写字母
// - 至少 1 个数字
// - 至少 1 个特殊字符
func ValidatePasswordStrength(plaintext string) error {
	if len(plaintext) < 8 {
		return errors.ErrInvalidPassword
	}

	var hasUpper, hasLower, hasDigit bool
	for _, c := range plaintext {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}

	hasSpecial := passwordSpecialChars.MatchString(plaintext)

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return errors.ErrInvalidPassword
	}
	return nil
}
