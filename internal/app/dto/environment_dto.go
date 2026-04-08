package dto

import (
	"time"
)

type CreateEnvironmentRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Key         string `json:"key" binding:"required,max=20"`
	Description string `json:"description" binding:"max=200"`
	Color       string `json:"color" binding:"omitempty,hexcolor"`
}

type UpdateEnvironmentRequest struct {
	Name        string `json:"name" binding:"omitempty,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
	Color       string `json:"color" binding:"omitempty,hexcolor"`
}

type EnvironmentResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Key          string    `json:"key"`
	Description  string    `json:"description"`
	Color        string    `json:"color"`
	ClusterCount int       `json:"clusterCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type ListEnvironmentsRequest struct {
	Page     int `json:"page" binding:"min=1"`
	PageSize int `json:"pageSize" binding:"min=1,max=100"`
}

type ListEnvironmentsResponse struct {
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
	List     []EnvironmentResponse `json:"list"`
}
