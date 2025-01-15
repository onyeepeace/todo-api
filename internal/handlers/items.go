package handlers

import (
	"encoding/json"
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

	rows, err := db.Query("SELECT item_id, user_id, name, content, created_at, updated_at FROM items WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to retrieve items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ItemID, &item.UserID, &item.Name, &item.Content, &item.CreatedAt, &item.UpdatedAt); err != nil {
			http.Error(w, "Failed to scan item", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	var item models.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	err := db.QueryRow(
		"INSERT INTO items (user_id, name, content) VALUES ($1, $2, $3) RETURNING item_id, user_id, name, content, created_at, updated_at",
		userID, item.Name, item.Content,
	).Scan(&item.ItemID, &item.UserID, &item.Name, &item.Content, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		http.Error(w, "Failed to insert item", http.StatusInternalServerError)
		return
	}

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

	query := `SELECT item_id, user_id, name, content, created_at, updated_at FROM items WHERE item_id = $1`
	row := db.QueryRow(query, itemID)

	var item models.Item
	if err := row.Scan(&item.ItemID, &item.UserID, &item.Name, &item.Content, &item.CreatedAt, &item.UpdatedAt); err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func EditItemHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var updatedItem models.Item
	if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(models.UserIDKey).(int)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	query := `UPDATE items SET name = $1, content = $2, updated_at = NOW() WHERE item_id = $3 AND user_id = $4 RETURNING item_id, user_id, name, content, created_at, updated_at`
	row := db.QueryRow(query, updatedItem.Name, updatedItem.Content, itemID, userID)

	var item models.Item
	if err := row.Scan(&item.ItemID, &item.UserID, &item.Name, &item.Content, &item.CreatedAt, &item.UpdatedAt); err != nil {
		http.Error(w, "Item not found or update failed", http.StatusNotFound)
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
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var shareRequest struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&shareRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if shareRequest.Role != "Viewer" && shareRequest.Role != "Collaborator" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO shared_items (item_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT (item_id, user_id) DO UPDATE SET role = $3", itemID, shareRequest.UserID, shareRequest.Role)
	if err != nil {
		http.Error(w, "Failed to share item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}