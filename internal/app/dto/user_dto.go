package dto

// CreateUserRequest 创建用户请求 DTO
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// UpdateUserRequest 更新用户请求 DTO
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" binding:"omitempty,min=2,max=50"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
}

// UserResponse 用户响应 DTO
type UserResponse struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListUsersRequest 用户列表请求 DTO
type ListUsersRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// ListUsersResponse 用户列表响应 DTO
type ListUsersResponse struct {
	Items    []*UserResponse `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}
