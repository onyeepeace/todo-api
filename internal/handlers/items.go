package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/models"
)

func GetItemsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	// Get all items where user has any role (owner, editor, or viewer)
	query := `
		SELECT DISTINCT 
			i.item_id,
			i.name,
			i.content,
			i.created_at,
			i.updated_at,
			r.name as role_name,
			u.email as shared_by_email
		FROM items i
		JOIN user_roles ur ON i.item_id = ur.item_id
		JOIN roles r ON ur.role_id = r.role_id
		LEFT JOIN users u ON ur.created_by = u.user_id
		WHERE ur.user_id = $1
		ORDER BY i.created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("Error retrieving items: %v", err)
		http.Error(w, "Failed to retrieve items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ItemWithAccess struct {
		models.Item
		Role      string `json:"role"`       // owner, editor, or viewer
		SharedBy  string `json:"shared_by"`  // email of user who shared it (null if owner)
	}

	var items []ItemWithAccess
	for rows.Next() {
		var item ItemWithAccess
		var sharedByEmail sql.NullString // Use sql.NullString for potentially null shared_by_email

		if err := rows.Scan(
			&item.ItemID,
			&item.Name,
			&item.Content,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Role,
			&sharedByEmail,
		); err != nil {
			log.Printf("Error scanning item: %v", err)
			http.Error(w, "Failed to scan item", http.StatusInternalServerError)
			return
		}

		// Only set SharedBy if the item was shared (role is not owner)
		if item.Role != "owner" && sharedByEmail.Valid {
			item.SharedBy = sharedByEmail.String
		}

		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		log.Printf("Error: User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	var item models.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure content is valid JSON
	if item.Content == nil {
		item.Content = json.RawMessage("[]")
	}

	// Start transaction
	tx, err := db.DB().Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Rollback if we don't commit

	// Create item
	err = tx.QueryRow(
		"INSERT INTO items (name, content) VALUES ($1, $2::jsonb) RETURNING item_id, name, content, created_at, updated_at",
		item.Name, item.Content,
	).Scan(&item.ItemID, &item.Name, &item.Content, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		log.Printf("Error creating item: %v", err)
		http.Error(w, "Failed to create item", http.StatusInternalServerError)
		return
	}

	// Get owner role ID
	var ownerRoleID int
	err = tx.QueryRow("SELECT role_id FROM roles WHERE name = 'owner'").Scan(&ownerRoleID)
	if err != nil {
		log.Printf("Error getting owner role: %v", err)
		http.Error(w, "Failed to get owner role", http.StatusInternalServerError)
		return
	}

	// Assign owner role to creator
	_, err = tx.Exec(
		"INSERT INTO user_roles (item_id, user_id, role_id, created_by) VALUES ($1, $2, $3, $2)",
		item.ItemID, userID, ownerRoleID,
	)
	if err != nil {
		log.Printf("Error assigning owner role: %v", err)
		http.Error(w, "Failed to assign owner role", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func GetItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Query item with user's role and who shared it
	query := `
		SELECT 
			i.item_id,
			i.name,
			i.content,
			i.created_at,
			i.updated_at,
			r.name as role_name,
			u.email as shared_by_email
		FROM items i
		JOIN user_roles ur ON i.item_id = ur.item_id
		JOIN roles r ON ur.role_id = r.role_id
		LEFT JOIN users u ON ur.created_by = u.user_id
		WHERE i.item_id = $1 AND ur.user_id = $2
	`
	row := db.QueryRow(query, itemID, userID)

	type ItemWithAccess struct {
		models.Item
		Role     string `json:"role"`      // owner, editor, or viewer
		SharedBy string `json:"shared_by"` // email of user who shared it (null if owner)
	}

	var item ItemWithAccess
	var sharedByEmail sql.NullString

	err = row.Scan(
		&item.ItemID,
		&item.Name,
		&item.Content,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.Role,
		&sharedByEmail,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
		} else {
			log.Printf("Error scanning item: %v", err)
			http.Error(w, "Failed to get item", http.StatusInternalServerError)
		}
		return
	}

	// Only set SharedBy if the item was shared (role is not owner)
	if item.Role != "owner" && sharedByEmail.Valid {
		item.SharedBy = sharedByEmail.String
	}

	// Generate ETag
	etag := item.Item.GenerateETag()
	w.Header().Set("ETag", etag)

	// Check If-None-Match header for cache validation
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

type UpdateItemRequest struct {
	Name    string          `json:"name"`
	Content json.RawMessage `json:"content"`
}

func EditItemHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var updateReq UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.DB().Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get current item state
	var currentItem models.Item
	err = tx.QueryRow(`
		SELECT item_id, name, content, created_at, updated_at
		FROM items WHERE item_id = $1
	`, itemID).Scan(
		&currentItem.ItemID,
		&currentItem.Name,
		&currentItem.Content,
		&currentItem.CreatedAt,
		&currentItem.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error getting current item state: %v", err)
		http.Error(w, "Failed to get current item state", http.StatusInternalServerError)
		return
	}

	// Check If-Match header
	if match := r.Header.Get("If-Match"); match != "" {
		if !currentItem.ValidateETag(match) {
			http.Error(w, "Precondition Failed - Item has been modified", http.StatusPreconditionFailed)
			return
		}
	}

	// Update the item
	var updatedItem models.Item
	err = tx.QueryRow(`
		UPDATE items 
		SET name = $1, content = $2::jsonb, updated_at = NOW()
		WHERE item_id = $3 
		RETURNING item_id, name, content, created_at, updated_at
	`, updateReq.Name, updateReq.Content, itemID).Scan(
		&updatedItem.ItemID,
		&updatedItem.Name,
		&updatedItem.Content,
		&updatedItem.CreatedAt,
		&updatedItem.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error updating item: %v", err)
		http.Error(w, "Failed to update item", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Generate new ETag
	etag := updatedItem.GenerateETag()
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedItem)
}

func DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.DB().Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete associated todos first
	_, err = tx.Exec("DELETE FROM todos WHERE item_id = $1", itemID)
	if err != nil {
		log.Printf("Error deleting todos: %v", err)
		http.Error(w, "Failed to delete todos", http.StatusInternalServerError)
		return
	}

	// Delete user roles
	_, err = tx.Exec("DELETE FROM user_roles WHERE item_id = $1", itemID)
	if err != nil {
		log.Printf("Error deleting user roles: %v", err)
		http.Error(w, "Failed to delete user roles", http.StatusInternalServerError)
		return
	}

	// Delete the item
	result, err := tx.Exec("DELETE FROM items WHERE item_id = $1", itemID)
	if err != nil {
		log.Printf("Error deleting item: %v", err)
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ShareItemHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var shareRequest struct {
		UserID int    `json:"user_id"` // ID of user to share with
		Role   string `json:"role"`    // editor or viewer
	}
	if err := json.NewDecoder(r.Body).Decode(&shareRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if shareRequest.UserID == 0 {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	if shareRequest.Role == "" {
		http.Error(w, "role is required", http.StatusBadRequest)
		return
	}

	// Validate role
	if shareRequest.Role != "editor" && shareRequest.Role != "viewer" {
		http.Error(w, "Invalid role. Must be 'editor' or 'viewer'", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to share item %d with user ID: %d, role: %s", itemID, shareRequest.UserID, shareRequest.Role)

	// Start transaction
	tx, err := db.DB().Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Verify user exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", shareRequest.UserID).Scan(&exists)
	if err != nil {
		log.Printf("Error verifying user existence: %v", err)
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get role ID
	var roleID int
	err = tx.QueryRow("SELECT role_id FROM roles WHERE name = $1", shareRequest.Role).Scan(&roleID)
	if err != nil {
		log.Printf("Error getting role: %v", err)
		http.Error(w, "Failed to get role", http.StatusInternalServerError)
		return
	}

	log.Printf("Found role ID: %d for role: %s", roleID, shareRequest.Role)

	// Check if user already has a role for this item
	var existingRoleID int
	err = tx.QueryRow("SELECT role_id FROM user_roles WHERE user_id = $1 AND item_id = $2", shareRequest.UserID, itemID).Scan(&existingRoleID)
	if err == nil {
		// Update existing role
		_, err = tx.Exec(
			"UPDATE user_roles SET role_id = $1, created_by = $2 WHERE user_id = $3 AND item_id = $4",
			roleID, userID, shareRequest.UserID, itemID,
		)
		log.Printf("Updated existing role for user")
	} else if err == sql.ErrNoRows {
		// Insert new role
		_, err = tx.Exec(
			"INSERT INTO user_roles (item_id, user_id, role_id, created_by) VALUES ($1, $2, $3, $4)",
			itemID, shareRequest.UserID, roleID, userID,
		)
		log.Printf("Inserted new role for user")
	}

	if err != nil {
		log.Printf("Error updating user role: %v", err)
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}