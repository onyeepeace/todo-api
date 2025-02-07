package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

var db *sql.DB

// Initialize sets up the database connection and creates tables
func Initialize(config Config) (*sql.DB, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Create connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	var err error
	// Connect to database
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	log.Printf("Successfully connected to database")

	// Create tables and initial data
	if err := createSchema(); err != nil {
		return nil, fmt.Errorf("error creating schema: %v", err)
	}

	return db, nil
}

// Query executes a query that returns rows
func Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(query, args...)
}

// QueryRow executes a query that returns a single row
func QueryRow(query string, args ...interface{}) *sql.Row {
	return db.QueryRow(query, args...)
}

// Exec executes a query that doesn't return rows
func Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.Exec(query, args...)
}

// DB returns the database instance
func DB() *sql.DB {
	return db
}

// Close closes the database connection
func Close() error {
	return db.Close()
}

// createSchema creates all tables and initial data
func createSchema() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create tables
	tables := `
		CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			provider_user_id VARCHAR(255) NOT NULL UNIQUE,
			provider VARCHAR(50) NOT NULL DEFAULT 'google',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS roles (
			role_id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS permissions (
			permission_id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS role_permissions (
			role_id INT REFERENCES roles(role_id) ON DELETE CASCADE,
			permission_id INT REFERENCES permissions(permission_id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (role_id, permission_id)
		);

		CREATE TABLE IF NOT EXISTS items (
			item_id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			content JSONB NOT NULL DEFAULT '{}',
			version INT NOT NULL DEFAULT 1,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS user_roles (
			item_id INT REFERENCES items(item_id) ON DELETE CASCADE,
			role_id INT REFERENCES roles(role_id) ON DELETE CASCADE,
			user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
			created_by INT REFERENCES users(user_id) ON DELETE SET NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (item_id, user_id)
		);

		CREATE TABLE IF NOT EXISTS todos (
			todo_id SERIAL PRIMARY KEY,
			item_id INT REFERENCES items(item_id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			done BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := tx.Exec(tables); err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	// Create trigger function for ensuring item ownership
	triggerFunc := `
		CREATE OR REPLACE FUNCTION ensure_item_owner()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Check if this is an owner role being removed
			IF EXISTS (
				SELECT 1 FROM roles 
				WHERE role_id = OLD.role_id 
				AND name = 'owner'
			) THEN
				-- Only allow owner to remove themselves
				IF OLD.user_id != CURRENT_USER_ID() THEN
					RAISE EXCEPTION 'Only the owner can remove themselves from an item';
				END IF;
			END IF;

			RETURN OLD;
		END;
		$$ LANGUAGE plpgsql;
	`

	if _, err := tx.Exec(triggerFunc); err != nil {
		return fmt.Errorf("error creating trigger function: %v", err)
	}

	// Create trigger without the WHEN clause
	trigger := `
		DROP TRIGGER IF EXISTS prevent_remove_last_owner ON user_roles;
		CREATE TRIGGER prevent_remove_last_owner
		BEFORE DELETE ON user_roles
		FOR EACH ROW
		EXECUTE FUNCTION ensure_item_owner();
	`

	if _, err := tx.Exec(trigger); err != nil {
		return fmt.Errorf("error creating trigger: %v", err)
	}

	// Insert initial data
	initialData := `
		INSERT INTO roles (name, description) VALUES
			('owner', 'Full control over the item and can manage other users'' access'),
			('editor', 'Can view and edit the item content'),
			('viewer', 'Can only view the item content')
		ON CONFLICT (name) DO NOTHING;

		INSERT INTO permissions (name, description) VALUES
			('can_view', 'Can view the item content'),
			('can_edit', 'Can modify the item content'),
			('can_share', 'Can share the item with other users'),
			('can_delete', 'Can delete the item')
		ON CONFLICT (name) DO NOTHING;

		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.role_id, p.permission_id
		FROM roles r, permissions p
		WHERE 
			(r.name = 'owner') -- owner gets all permissions
			OR (r.name = 'editor' AND p.name IN ('can_view', 'can_edit'))
			OR (r.name = 'viewer' AND p.name = 'can_view')
		ON CONFLICT DO NOTHING;
	`

	if _, err := tx.Exec(initialData); err != nil {
		return fmt.Errorf("error inserting initial data: %v", err)
	}

	return tx.Commit()
}
