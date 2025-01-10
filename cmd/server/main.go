package main

import (
	"log"
	"net/http"

	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/handlers"
	"github.com/onyeepeace/todo-api/internal/middleware"
)

func main() {
	// Connect to the database
	db.ConnectToDB()
	defer db.Close()

	// Create the todos table if it doesn't exist
	if err := db.CreateTables(); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	mux := http.NewServeMux()

	// Healthcheck endpoint
	mux.HandleFunc("/healthcheck", handlers.HealthCheckHandler)

	// OAuth2 endpoints
	mux.HandleFunc("/api/auth/login", handlers.LoginHandler)
	mux.HandleFunc("/api/auth/callback", handlers.CallbackHandler)
	mux.HandleFunc("/api/auth/logout", handlers.LogoutHandler)

	// Protected routes
	protectedMux := http.NewServeMux()

	// Lists endpoints
	protectedMux.HandleFunc("/api/lists", handlers.ListsHandler)
	protectedMux.HandleFunc("/api/lists/", handlers.ListByIDHandler)
	protectedMux.HandleFunc("/api/lists/share", handlers.ShareListHandler)

	// Todos endpoints
	protectedMux.HandleFunc("/api/lists/{list_id}/todos", handlers.TodosHandler)
	protectedMux.HandleFunc("/api/lists/{list_id}/todos/", handlers.TodoByIDHandler)

	// Wrap protectedMux with JWT validation middleware
	protectedHandler := middleware.ValidateJWT(protectedMux.ServeHTTP)

	// Mount protectedMux under /api
	mux.Handle("/api/", protectedHandler)

	// Start server with CORS middleware
	log.Fatal(http.ListenAndServe(":4000", middleware.CorsMiddleware(mux)))
}