package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/onyeepeace/todo-api/internal/models"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

// Authorize middleware checks if the user has the required permission
func Authorize(db *sql.DB, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(models.UserIDKey).(int)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract item_id from URL parameters
			itemIDStr := chi.URLParam(r, "item_id")
			if itemIDStr == "" {
				// If no item_id in URL, this might be a list endpoint or creation endpoint
				// Let it pass through as the handler will handle appropriate checks
				next.ServeHTTP(w, r)
				return
			}

			itemID, err := strconv.Atoi(itemIDStr)
			if err != nil {
				http.Error(w, "Invalid item ID", http.StatusBadRequest)
				return
			}

			// Check if user has the required permission
			var exists bool
			err = db.QueryRow(`
				SELECT EXISTS (
					SELECT 1 FROM user_roles ur
					JOIN roles r ON ur.role_id = r.role_id
					JOIN role_permissions rp ON r.role_id = rp.role_id
					JOIN permissions p ON rp.permission_id = p.permission_id
					WHERE ur.user_id = $1
					AND ur.item_id = $2
					AND p.name = $3
				)
			`, userID, itemID, requiredPermission).Scan(&exists)

			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !exists {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserPermissions returns all permissions a user has for an item
func GetUserPermissions(db *sql.DB, userID, itemID int) ([]string, error) {
	rows, err := db.Query(`
		SELECT DISTINCT p.name 
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.role_id
		JOIN role_permissions rp ON r.role_id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.permission_id
		WHERE ur.user_id = $1 AND ur.item_id = $2
	`, userID, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// CheckPermission checks if a user has a specific permission for an item
func CheckPermission(db *sql.DB, userID, itemID int, permission string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur
			JOIN roles r ON ur.role_id = r.role_id
			JOIN role_permissions rp ON r.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.permission_id
			WHERE ur.user_id = $1
			AND ur.item_id = $2
			AND p.name = $3
		)
	`, userID, itemID, permission).Scan(&exists)

	return exists, err
}