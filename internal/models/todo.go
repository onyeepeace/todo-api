package models

import "time"

type Todo struct {
	TodoID    	int       	`json:"todo_id"`
	ItemID 		int    		`json:"item_id"`
	Title 		string 		`json:"title"`
	Done  		bool   		`json:"done"`
	CreatedAt 	time.Time 	`json:"created_at"`
	UpdatedAt 	time.Time 	`json:"updated_at"`
}
