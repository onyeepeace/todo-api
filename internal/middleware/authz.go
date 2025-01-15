package middleware

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/models"
)

func Authorize(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(models.UserIDKey).(int)
			if !ok {
				http.Error(w, "User ID not found in context", http.StatusInternalServerError)
				return
			}

			if chi.URLParam(r, "item_id") != "" {
				itemIDStr := chi.URLParam(r, "item_id")
				itemID, err := strconv.Atoi(itemIDStr)
				if err != nil {
					http.Error(w, "Invalid item ID", http.StatusBadRequest)
					return
				}

				if !HasPermission(userID, itemID, action) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func HasPermission(userID, itemID int, action string) bool {
	var role string
	err := db.QueryRow("SELECT role FROM shared_items WHERE user_id = $1 AND item_id = $2", userID, itemID).Scan(&role)
	if err != nil {
		return false
	}

	switch action {
	case "edit":
		return role == "Owner" || role == "Collaborator"
	case "delete":
		return role == "Owner"
	case "view", "view_todos":
		return role == "Owner" || role == "Collaborator" || role == "Viewer"
	case "add_todo", "edit_todo", "delete_todo":
		return role == "Owner" || role == "Collaborator"
	default:
		return false
	}
} 