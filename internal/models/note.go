package models

type Note struct {
	NoteID    string `json:"note_id"`
	ItemID    int    `json:"item_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
