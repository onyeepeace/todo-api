package models

type Todo struct {
	TodoID    int    `json:"todo_id"`
	ListID int    `json:"list_id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Body  string `json:"body"`
}
