package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type Item struct {
	ItemID    int             `json:"item_id"`
	Name      string          `json:"name"`
	Content   json.RawMessage `json:"content"`
	ETag      string          `json:"etag,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// GenerateETag creates a hash of the item's content and metadata
func (i *Item) GenerateETag() string {
	// Combine relevant fields into a string
	content := fmt.Sprintf("%d-%s-%s-%s",
		i.ItemID,
		i.Name,
		string(i.Content),
		i.UpdatedAt.Format(time.RFC3339Nano),
	)

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(content))
	
	// Return as hex string with quotes as per HTTP ETag spec
	return fmt.Sprintf(`"%x"`, hash[:])
}

// ValidateETag checks if the provided ETag matches the item's current state
func (i *Item) ValidateETag(etag string) bool {
	if etag == "" {
		return false
	}
	return i.GenerateETag() == etag
}
