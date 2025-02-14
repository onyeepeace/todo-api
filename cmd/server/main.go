package main

import (
	"log"
	"net/http"
	"os"
	"strings"

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
	
	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			allowedOriginsMap := make(map[string]bool)
			for _, o := range allowedOrigins {
				allowedOriginsMap[strings.TrimSpace(o)] = true
			}
			return allowedOriginsMap[origin]
		},
	}))

	r.Get("/healthcheck", handlers.HealthCheckHandler)

	r.Route("/api/auth", func(r chi.Router) {
		r.Get("/callback", handlers.CallbackHandler)
		r.Get("/logout", handlers.LogoutHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.ValidateJWT)

		r.Route("/api/items", func(r chi.Router) {
			r.Get("/", handlers.GetItemsHandler)
			r.Post("/", handlers.CreateItemHandler)
			
			// Routes that need item_id
			r.Group(func(r chi.Router) {
				r.With(middleware.Authorize(db.DB(), "can_view")).Get("/{item_id}", handlers.GetItemByIDHandler)
				r.With(middleware.Authorize(db.DB(), "can_edit")).Put("/{item_id}", handlers.EditItemHandler)
				r.With(middleware.Authorize(db.DB(), "can_edit")).Delete("/{item_id}", handlers.DeleteItemHandler)
				r.With(middleware.Authorize(db.DB(), "can_edit")).Post("/{item_id}/share", handlers.ShareItemHandler)

				// Todos routes
				r.Route("/{item_id}/todos", func(r chi.Router) {
					r.With(middleware.Authorize(db.DB(), "can_view")).Get("/", handlers.GetTodosHandler)
					r.With(middleware.Authorize(db.DB(), "can_edit")).Post("/", handlers.CreateTodoHandler)
					r.With(middleware.Authorize(db.DB(), "can_view")).Get("/{todo_id}", handlers.GetTodoByIDHandler)
					r.With(middleware.Authorize(db.DB(), "can_edit")).Put("/{todo_id}", handlers.EditTodoHandler)
					r.With(middleware.Authorize(db.DB(), "can_edit")).Patch("/{todo_id}/done", handlers.MarkTodoDoneHandler)
					r.With(middleware.Authorize(db.DB(), "can_edit")).Delete("/{todo_id}", handlers.DeleteTodoHandler)
				})
			})
		})

		// Add users endpoints
		r.Route("/api/users", func(r chi.Router) {
			r.Get("/lookup", handlers.LookupUserHandler)
			r.Get("/me", handlers.GetCurrentUserHandler)
		})
	})

	log.Fatal(http.ListenAndServe(":4000", r))
}