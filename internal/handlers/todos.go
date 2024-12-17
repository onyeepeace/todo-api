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

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://todo-app-go.vercel.app")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TodosHandler(w http.ResponseWriter, r *http.Request) {
	listID, err := parseListIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getTodosHandler(w, r, listID)
	case http.MethodPost:
		createTodoHandler(w, r, listID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func TodoByIDHandler(w http.ResponseWriter, r *http.Request) {
	listID, todoID, err := parseListAndTodoIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid list or todo ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		markTodoDoneHandler(w, r, listID, todoID)
	case http.MethodPut:
		editTodoHandler(w, r, listID, todoID)
	case http.MethodDelete:
		deleteTodoHandler(w, r, listID, todoID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTodosHandler(w http.ResponseWriter, r *http.Request, listID int) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT todo_id, title, body, done FROM todos WHERE list_id = $1", listID)
	if err != nil {
		http.Error(w, "Failed to retrieve todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.TodoID, &todo.Title, &todo.Body, &todo.Done); err != nil {
			http.Error(w, "Failed to scan todo", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func createTodoHandler(w http.ResponseWriter, r *http.Request, listID int) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.QueryRow(
		"INSERT INTO todos (list_id, title, body, done) VALUES ($1, $2, $3, $4) RETURNING todo_id",
		listID, todo.Title, todo.Body, todo.Done,
	).Scan(&todo.TodoID)
	if err != nil {
		http.Error(w, "Failed to insert todo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func markTodoDoneHandler(w http.ResponseWriter, r *http.Request, listID, todoID int) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := db.Exec("UPDATE todos SET done = true WHERE list_id = $1 AND todo_id = $2", listID, todoID)
	if err != nil {
		http.Error(w, "Failed to mark todo as done", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func editTodoHandler(w http.ResponseWriter, r *http.Request, listID, todoID int) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}


	var updatedTodo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&updatedTodo); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `UPDATE todos SET title = $1, body = $2, done = $3 
              WHERE list_id = $4 AND todo_id = $5 
              RETURNING todo_id, title, body, done`
	row := db.QueryRow(query, updatedTodo.Title, updatedTodo.Body, updatedTodo.Done, listID, todoID)

	var todo models.Todo
	if err := row.Scan(&todo.TodoID, &todo.Title, &todo.Body, &todo.Done); err != nil {
		http.Error(w, "Todo not found or update failed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func deleteTodoHandler(w http.ResponseWriter, r *http.Request, listID, todoID int) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `DELETE FROM todos WHERE list_id = $1 AND todo_id = $2`
	result, err := db.Exec(query, listID, todoID)
	if err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Parsing helpers
func parseListIDFromPath(path string) (int, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[0] != "api" || parts[1] != "lists" {
		return 0, errors.New("invalid path format")
	}
	return strconv.Atoi(parts[2])
}

func parseListAndTodoIDFromPath(path string) (int, int, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 || parts[0] != "api" || parts[1] != "lists" || parts[3] != "todos" {
		return 0, 0, errors.New("invalid path format")
	}
	listID, err1 := strconv.Atoi(parts[2])
	todoID, err2 := strconv.Atoi(parts[4])
	if err1 != nil || err2 != nil {
		return 0, 0, errors.New("invalid list or todo ID")
	}
	return listID, todoID, nil
}