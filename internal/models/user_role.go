package models

import "time"

type UserRole struct {
	ItemID    int       `json:"item_id"`
	RoleID    int       `json:"role_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy int       `json:"created_by"`
} 