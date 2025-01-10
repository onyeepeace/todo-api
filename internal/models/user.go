package models

import "time"

type User struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	ProviderUserID string   `json:"provider_user_id"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}