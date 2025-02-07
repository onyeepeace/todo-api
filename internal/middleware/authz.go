package middleware

import (
	"database/sql"
	"errors"
	"net/http"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

// CheckUserPermission middleware checks if the user has the required permission for an item
func CheckUserPermission(db *sql.DB, requiredPermission string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("user_id").(int) // Set by auth middleware
			itemID := r.Context().Value("item_id").(int) // Set by previous middleware

			// Check if user has the required permission
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
		}
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

// Authorize middleware checks if the user has the required permission
func Authorize(db *sql.DB, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("user_id").(int)
			itemID := r.Context().Value("item_id").(int)

			hasPermission, err := CheckPermission(db, userID, itemID, action)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}