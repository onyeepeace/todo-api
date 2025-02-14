package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/models"
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

func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var user models.User
	err := db.QueryRow(`
		SELECT user_id, email, username, created_at 
		FROM users 
		WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Email, &user.Username, &user.CreatedAt)

	if err != nil {
		log.Printf("Error getting user details: %v", err)
		http.Error(w, "Failed to get user details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
} 