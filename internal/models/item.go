package models

import (
	"encoding/json"
	"time"
)

type Item struct {
	ItemID    int             `json:"item_id"`
	UserID    int             `json:"user_id"`
	Name      string          `json:"name"`
	Content   json.RawMessage `json:"content"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}