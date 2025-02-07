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
			i.version,
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
			&item.Version,
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
	// Get user ID from context
	userID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		log.Printf("Error: User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	// Decode request body
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

	log.Printf("Creating item with name: %s and content: %s", item.Name, string(item.Content))

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
		"INSERT INTO items (name, content) VALUES ($1, $2::jsonb) RETURNING item_id, name, content, version, created_at, updated_at",
		item.Name, item.Content,
	).Scan(&item.ItemID, &item.Name, &item.Content, &item.Version, &item.CreatedAt, &item.UpdatedAt)
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

	log.Printf("Found owner role ID: %d", ownerRoleID)

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

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created item with ID: %d", item.ItemID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func GetItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	query := `SELECT item_id, name, content, version, created_at, updated_at FROM items WHERE item_id = $1`
	row := db.QueryRow(query, itemID)

	var item models.Item
	if err := row.Scan(&item.ItemID, &item.Name, &item.Content, &item.Version, &item.CreatedAt, &item.UpdatedAt); err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

type UpdateItemRequest struct {
	Name    string          `json:"name"`
	Content json.RawMessage `json:"content"`
	Version int            `json:"version"` // Client must send current version
}

func EditItemHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	// Get item ID from URL
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var updateReq UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to update item %d with version %d", itemID, updateReq.Version)

	// Start transaction
	tx, err := db.DB().Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get current version
	var currentVersion int
	err = tx.QueryRow("SELECT version FROM items WHERE item_id = $1", itemID).Scan(&currentVersion)
	if err != nil {
		log.Printf("Error getting current version: %v", err)
		http.Error(w, "Failed to get current version", http.StatusInternalServerError)
		return
	}

	log.Printf("Current version in DB: %d, Client version: %d", currentVersion, updateReq.Version)

	// Check if user has edit permission
	var canEdit bool
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur
			JOIN roles r ON ur.role_id = r.role_id
			JOIN role_permissions rp ON r.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.permission_id
			WHERE ur.user_id = $1
			AND ur.item_id = $2
			AND p.name = 'can_edit'
		)
	`, userID, itemID).Scan(&canEdit)

	if err != nil {
		log.Printf("Error checking edit permission: %v", err)
		http.Error(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !canEdit {
		http.Error(w, "You don't have permission to edit this item", http.StatusForbidden)
		return
	}

	// Update the item with version check
	var item models.Item
	err = tx.QueryRow(`
		UPDATE items 
		SET name = $1, 
			content = $2, 
			version = version + 1,
			updated_at = NOW() 
		WHERE item_id = $3 
		AND version = $4
		RETURNING item_id, name, content, version, created_at, updated_at
	`, updateReq.Name, updateReq.Content, itemID, updateReq.Version).Scan(
		&item.ItemID,
		&item.Name,
		&item.Content,
		&item.Version,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Check if item exists but version doesn't match
			var exists bool
			err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM items WHERE item_id = $1)", itemID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking item existence: %v", err)
				http.Error(w, "Failed to verify item existence", http.StatusInternalServerError)
				return
			}
			if exists {
				log.Printf("Version mismatch detected. Update failed.")
				http.Error(w, "Item was modified by another user. Please refresh and try again.", http.StatusConflict)
			} else {
				http.Error(w, "Item not found", http.StatusNotFound)
			}
		} else {
			log.Printf("Error updating item: %v", err)
			http.Error(w, "Failed to update item", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Successfully updated item %d to version %d", itemID, item.Version)

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to update item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM items WHERE item_id = $1`
	result, err := db.Exec(query, itemID)
	if err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ShareItemHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user ID (the sharer)
	currentUserID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	// Get item ID from URL
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

	// Validate role
	if shareRequest.Role != "editor" && shareRequest.Role != "viewer" {
		http.Error(w, "Invalid role. Must be 'editor' or 'viewer'", http.StatusBadRequest)
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

	// Verify current user has permission to share
	var canShare bool
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur
			JOIN roles r ON ur.role_id = r.role_id
			JOIN role_permissions rp ON r.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.permission_id
			WHERE ur.user_id = $1
			AND ur.item_id = $2
			AND p.name = 'can_share'
		)
	`, currentUserID, itemID).Scan(&canShare)

	if err != nil {
		log.Printf("Error checking share permission: %v", err)
		http.Error(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !canShare {
		http.Error(w, "You don't have permission to share this item", http.StatusForbidden)
		return
	}

	// Get role ID for the requested role
	var roleID int
	err = tx.QueryRow("SELECT role_id FROM roles WHERE name = $1", shareRequest.Role).Scan(&roleID)
	if err != nil {
		log.Printf("Error getting role ID: %v", err)
		http.Error(w, "Invalid role", http.StatusInternalServerError)
		return
	}

	// Share the item (upsert in case the user already has a role)
	_, err = tx.Exec(`
		INSERT INTO user_roles (item_id, user_id, role_id, created_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (item_id, user_id) 
		DO UPDATE SET role_id = $3, created_by = $4
	`, itemID, shareRequest.UserID, roleID, currentUserID)

	if err != nil {
		log.Printf("Error sharing item: %v", err)
		http.Error(w, "Failed to share item", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to share item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}