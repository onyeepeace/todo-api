package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/models"
)

func GetTodosHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT todo_id, title, done FROM todos WHERE item_id = $1", itemID)
	if err != nil {
		http.Error(w, "Failed to retrieve todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.TodoID, &todo.Title, &todo.Done); err != nil {
			http.Error(w, "Failed to scan todo", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = db.QueryRow(
		"INSERT INTO todos (item_id, title, done) VALUES ($1, $2, $3) RETURNING todo_id",
		itemID, todo.Title, todo.Done,
	).Scan(&todo.TodoID)
	if err != nil {
		http.Error(w, "Failed to insert todo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func GetTodoByIDHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	todoIDStr := chi.URLParam(r, "todo_id")
	itemID, err1 := strconv.Atoi(itemIDStr)
	todoID, err2 := strconv.Atoi(todoIDStr)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid item or todo ID", http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT todo_id, title, done FROM todos WHERE item_id = $1 AND todo_id = $2", itemID, todoID)

	var todo models.Todo
	if err := row.Scan(&todo.TodoID, &todo.Title, &todo.Done); err != nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func EditTodoHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	todoIDStr := chi.URLParam(r, "todo_id")
	itemID, err1 := strconv.Atoi(itemIDStr)
	todoID, err2 := strconv.Atoi(todoIDStr)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid item or todo ID", http.StatusBadRequest)
		return
	}

	var updatedTodo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&updatedTodo); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `UPDATE todos SET title = $1, done = $2 
              WHERE item_id = $3 AND todo_id = $4 
              RETURNING todo_id, title, done`
	row := db.QueryRow(query, updatedTodo.Title, updatedTodo.Done, itemID, todoID)

	var todo models.Todo
	if err := row.Scan(&todo.TodoID, &todo.Title, &todo.Done); err != nil {
		http.Error(w, "Todo not found or update failed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	todoIDStr := chi.URLParam(r, "todo_id")
	itemID, err1 := strconv.Atoi(itemIDStr)
	todoID, err2 := strconv.Atoi(todoIDStr)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid item or todo ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM todos WHERE item_id = $1 AND todo_id = $2`
	result, err := db.Exec(query, itemID, todoID)
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

func MarkTodoDoneHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	todoIDStr := chi.URLParam(r, "todo_id")
	itemID, err1 := strconv.Atoi(itemIDStr)
	todoID, err2 := strconv.Atoi(todoIDStr)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid item or todo ID", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE todos SET done = true WHERE item_id = $1 AND todo_id = $2", itemID, todoID)
	if err != nil {
		http.Error(w, "Failed to mark todo as done", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}