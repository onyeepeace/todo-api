package models

type Todo struct {
	TodoID    int    `json:"todo_id"`
	ItemID int    `json:"item_id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Body  string `json:"body"`
}
