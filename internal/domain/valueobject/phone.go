package valueobject

import (
	"mengri-flow/internal/domain/errors"
	"regexp"
	"strings"
)

// Phone 手机号值对象，不可变，自带格式校验。
type Phone struct {
	number string
}

var phoneRegex = regexp.MustCompile(`^\+\d{7,15}$`)

// NewPhone 创建并校验手机号值对象（格式：+86...）。
func NewPhone(raw string) (Phone, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Phone{}, errors.ErrNotFound
	}
	if !phoneRegex.MatchString(raw) {
		return Phone{}, errors.ErrNotFound
	}
	return Phone{number: raw}, nil
}

// String 返回完整手机号。
func (p Phone) String() string {
	return p.number
}

// Masked 返回脱敏手机号（如 +861380****000）。
func (p Phone) Masked() string {
	n := p.number
	if len(n) <= 7 {
		return n
	}
	// 保留前 4 位和后 3 位
	return n[:4] + "****" + n[len(n)-3:]
}

// Equals 值对象相等性比较。
func (p Phone) Equals(other Phone) bool {
	return p.number == other.number
}
