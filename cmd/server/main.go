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
	
	mux.HandleFunc("/api/lists", handlers.ListsHandler)
	mux.HandleFunc("/api/lists/", handlers.ListByIDHandler)
	
	mux.HandleFunc("/api/lists/{list_id}/todos", handlers.TodosHandler)
	mux.HandleFunc("/api/lists/{list_id}/todos/", handlers.TodoByIDHandler)

	// Start server with CORS middleware
	log.Fatal(http.ListenAndServe(":4000", handlers.CorsMiddleware(mux)))
}