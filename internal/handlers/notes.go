package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/models"
)

func GetNotesHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT note_id, content FROM notes WHERE note_id = $1", itemID)
	if err != nil {
		http.Error(w, "Failed to retrieve notes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.NoteID, &note.Content); err != nil {
			http.Error(w, "Failed to scan note", http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var note models.Note
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = db.QueryRow(
		"INSERT INTO notes (note_id, content) VALUES ($1, $2) RETURNING note_id",
		itemID, note.Content,
	).Scan(&note.NoteID)
	if err != nil {
		http.Error(w, "Failed to insert note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func GetNoteByIDHandler(w http.ResponseWriter, r *http.Request) {
	noteIDStr := chi.URLParam(r, "note_id")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT note_id, content FROM notes WHERE note_id = $1", noteID)

	var note models.Note
	if err := row.Scan(&note.NoteID, &note.Content); err != nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func EditNoteHandler(w http.ResponseWriter, r *http.Request) {
	noteIDStr := chi.URLParam(r, "note_id")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	var updatedNote models.Note
	if err := json.NewDecoder(r.Body).Decode(&updatedNote); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `UPDATE notes SET content = $1, updated_at = NOW() WHERE note_id = $2 RETURNING note_id, content`
	row := db.QueryRow(query, updatedNote.Content, noteID)

	var note models.Note
	if err := row.Scan(&note.NoteID, &note.Content); err != nil {
		http.Error(w, "Note not found or update failed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	noteIDStr := chi.URLParam(r, "note_id")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM notes WHERE note_id = $1`
	result, err := db.Exec(query, noteID)
	if err != nil {
		http.Error(w, "Failed to delete note", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
} 