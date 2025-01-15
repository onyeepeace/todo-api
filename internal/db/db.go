package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func ConnectToDB() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Println("Warning: .env file not found.")
	}

	connStr := os.Getenv("DATABASE_URL")

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
	}

	log.Println("Connected to the database successfully!")
}

func Close() {
	db.Close()
}

func CreateTables() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			username VARCHAR(255) NOT NULL,
			provider_user_id VARCHAR(255) NOT NULL UNIQUE,
			provider VARCHAR(50) NOT NULL DEFAULT 'google',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS items (
			item_id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(user_id) ON DELETE CASCADE NOT NULL,
			name TEXT NOT NULL,
			content JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);

		CREATE TABLE IF NOT EXISTS refresh_tokens (
			token_id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
			token TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL
		);

		CREATE TABLE IF NOT EXISTS notes (
			note_id SERIAL PRIMARY KEY REFERENCES items(item_id),
			item_id INT REFERENCES items(item_id) ON DELETE CASCADE,
			content JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS todos (
			todo_id SERIAL PRIMARY KEY,
			item_id INT REFERENCES items(item_id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			body TEXT,
			done BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS shared_items (
			item_id INT REFERENCES items(item_id) ON DELETE CASCADE,
			user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL,
			PRIMARY KEY (item_id, user_id)
		);
	`)
	return err
}

func Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(query, args...)
}

func QueryRow(query string, args ...interface{}) *sql.Row {
	return db.QueryRow(query, args...)
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.Exec(query, args...)
}
