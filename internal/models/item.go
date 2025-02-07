package models

import (
	"encoding/json"
	"time"
)

type Item struct {
	ItemID    int             `json:"item_id"`
	Name      string          `json:"name"`
	Content   json.RawMessage `json:"content"`
	Version   int             `json:"version"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
