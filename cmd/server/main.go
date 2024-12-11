package main

import (
	"log"
	"net/http"

	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/handlers"
)

func main() {
	// Connect to the database
	db.ConnectToDB()
	defer db.Close()

	// Create the todos table if it doesn't exist
	if err := db.CreateTodosTable(); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	mux := http.NewServeMux()

	// Healthcheck endpoint
	mux.HandleFunc("/healthcheck", handlers.HealthCheckHandler)

	// Route for /api/todos
	mux.HandleFunc("/api/todos", handlers.TodosHandler)

	// Route for /api/todos/{id}
	mux.HandleFunc("/api/todos/", handlers.TodoByIDHandler)

	// Start server with CORS middleware
	log.Fatal(http.ListenAndServe(":4000", handlers.CorsMiddleware(mux)))
}