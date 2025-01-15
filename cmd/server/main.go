package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/handlers"
	"github.com/onyeepeace/todo-api/internal/middleware"
)

func main() {
	db.ConnectToDB()
	defer db.Close()

	if err := db.CreateTables(); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

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
			r.Route("/{item_id}/todos", func(r chi.Router) {
				r.Get("/", handlers.GetTodosHandler)
				r.Post("/", handlers.CreateTodoHandler)
				r.Get("/{todo_id}", handlers.GetTodoByIDHandler)
				r.Put("/{todo_id}", handlers.EditTodoHandler)
				r.Patch("/{todo_id}/done", handlers.MarkTodoDoneHandler)
				r.Delete("/{todo_id}", handlers.DeleteTodoHandler)
			})
			r.Route("/{item_id}/notes", func(r chi.Router) {
				r.Get("/", handlers.GetNotesHandler)
				r.Post("/", handlers.CreateNoteHandler)
				r.Get("/{note_id}", handlers.GetNoteByIDHandler)
				r.Put("/{note_id}", handlers.EditNoteHandler)
				r.Delete("/{note_id}", handlers.DeleteNoteHandler)
			})
		})

	})

	log.Fatal(http.ListenAndServe(":4000", r))
}