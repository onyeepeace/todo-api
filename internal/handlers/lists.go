package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/models"
)

func ListsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getListsHandler(w, r)
	case http.MethodPost:
		createListHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func ListByIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getListByIdHandler(w, r)
	case http.MethodPut:
		editListHandler(w, r)
	case http.MethodDelete:
		deleteListHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getListsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT list_id, name, created_at, updated_at FROM lists")
	if err != nil {
		http.Error(w, "Failed to retrieve lists", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var list models.List
		if err := rows.Scan(&list.ListID, &list.Name, &list.CreatedAt, &list.UpdatedAt); err != nil {
			http.Error(w, "Failed to scan list", http.StatusInternalServerError)
			return
		}
		lists = append(lists, list)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

func getListByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	list_id, err := parseIDFromListPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	query := `SELECT list_id, name, created_at, updated_at FROM lists WHERE list_id = $1`
	row := db.QueryRow(query, list_id)

	var list models.List
	if err := row.Scan(&list.ListID, &list.Name, &list.CreatedAt, &list.UpdatedAt); err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func createListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var list models.List
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.QueryRow(
		"INSERT INTO lists (name) VALUES ($1) RETURNING list_id",
		list.Name,
	).Scan(&list.ListID)
	if err != nil {
		http.Error(w, "Failed to insert list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func editListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	list_id, err := parseIDFromListPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedList models.List
	if err := json.NewDecoder(r.Body).Decode(&updatedList); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `UPDATE lists SET name = $1, updated_at = NOW() WHERE list_id = $2 RETURNING list_id, name, created_at, updated_at`
	row := db.QueryRow(query, updatedList.Name, list_id)

	var list models.List
	if err := row.Scan(&list.ListID, &list.Name, &list.CreatedAt, &list.UpdatedAt); err != nil {
		http.Error(w, "List not found or update failed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func deleteListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	list_id, err := parseIDFromListPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM lists WHERE list_id = $1`
	result, err := db.Exec(query, list_id)
	if err != nil {
		http.Error(w, "Failed to delete list", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseIDFromListPath(path string) (int, error) {
	idStr := strings.TrimPrefix(path, "/api/lists/")
	list_id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("invalid ID")
	}
	return list_id, nil
}

func ShareListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	listID, err := parseIDFromListPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	var shareRequest struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&shareRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO shared_lists (list_id, user_id) VALUES ($1, $2)", listID, shareRequest.UserID)
	if err != nil {
		http.Error(w, "Failed to share list", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}