package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/onyeepeace/todo-api/internal/db"
)

func LookupUserHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	var userID int
	err := db.QueryRow("SELECT user_id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
	})
} 