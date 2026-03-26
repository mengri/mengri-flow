package valueobject

import (
	"mengri-flow/internal/domain/errors"
	"regexp"
	"strings"
)

// Email 是一个值对象，不可变，自带校验逻辑。
type Email struct {
	address string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail 创建并校验 Email 值对象
func NewEmail(address string) (Email, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return Email{}, errors.ErrInvalidEmail
	}
	if !emailRegex.MatchString(address) {
		return Email{}, errors.ErrInvalidEmail
	}
	return Email{address: strings.ToLower(address)}, nil
}

// String 返回邮箱地址字符串
func (e Email) String() string {
	return e.address
}

// Equals 值对象相等性比较
func (e Email) Equals(other Email) bool {
	return e.address == other.address
}
