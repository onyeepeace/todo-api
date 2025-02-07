package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/handlers"
	"github.com/onyeepeace/todo-api/internal/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database
	config := db.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}

	if _, err := db.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthcheck", handlers.HealthCheckHandler)

	r.Route("/api/auth", func(r chi.Router) {
		r.Get("/login", handlers.LoginHandler)
		r.Get("/callback", handlers.CallbackHandler)
		r.Get("/logout", handlers.LogoutHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.ValidateJWT)

		r.Route("/api/items", func(r chi.Router) {
			r.Get("/", handlers.GetItemsHandler)
			r.Post("/", handlers.CreateItemHandler)
			r.Get("/{item_id}", handlers.GetItemByIDHandler)
			r.Put("/{item_id}", handlers.EditItemHandler)
			r.Delete("/{item_id}", handlers.DeleteItemHandler)
			r.Post("/{item_id}/share", handlers.ShareItemHandler)
			r.Route("/{item_id}/todos", func(r chi.Router) {
				r.Get("/", handlers.GetTodosHandler)
				r.Post("/", handlers.CreateTodoHandler)
				r.Get("/{todo_id}", handlers.GetTodoByIDHandler)
				r.Put("/{todo_id}", handlers.EditTodoHandler)
				r.Patch("/{todo_id}/done", handlers.MarkTodoDoneHandler)
				r.Delete("/{todo_id}", handlers.DeleteTodoHandler)
			})
		})

		// Add users endpoints
		r.Route("/api/users", func(r chi.Router) {
			r.Get("/lookup", handlers.LookupUserHandler)
		})
	})

	log.Fatal(http.ListenAndServe(":4000", r))
}