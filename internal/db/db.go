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
		log.Fatalf("Failed to connect to database %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database %v", err)
	}

	log.Println("Connected to the database successfully!")
}

func Close() {
	db.Close()
}

func CreateTodosTable() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS lists (
			list_id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS todos (
			todo_id SERIAL PRIMARY KEY,
			list_id INT REFERENCES lists(list_id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			body TEXT,
			done BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
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
