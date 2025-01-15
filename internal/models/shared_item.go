package models

type SharedItem struct {
    ItemID int    `json:"item_id"`
    UserID int    `json:"user_id"`
    Role   string `json:"role"`
} 