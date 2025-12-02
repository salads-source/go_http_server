package routes

import (
	"database/sql"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/salads-source/go_http_server/db"
	"github.com/salads-source/go_http_server/utils"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);`

	createEventsTable := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		location TEXT NOT NULL,
		dateTime DATETIME NOT NULL,
		user_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	createRegistrationsTable := `
	CREATE TABLE IF NOT EXISTS registrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER,
		user_id INTEGER,
		FOREIGN KEY(event_id) REFERENCES events(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	if _, err := testDB.Exec(createUsersTable); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	if _, err := testDB.Exec(createEventsTable); err != nil {
		t.Fatalf("Failed to create events table: %v", err)
	}

	if _, err := testDB.Exec(createRegistrationsTable); err != nil {
		t.Fatalf("Failed to create registrations table: %v", err)
	}

	return testDB
}

// setupTestRouter creates a Gin router with test database
func setupTestRouter(t *testing.T) (*sql.DB, *gin.Engine) {
	testDB := setupTestDB(t)

	// Replace the global DB with test DB
	originalDB := db.DB
	db.DB = testDB
	t.Cleanup(func() {
		testDB.Close()
		db.DB = originalDB
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router)

	return testDB, router
}

// generateTestToken creates a JWT token for testing
func generateTestToken(t *testing.T, email string, userId int64) string {
	token, err := utils.GenerateToken(email, userId)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}
	return token
}

// createTestUser creates a user in the test database and returns the user ID
func createTestUser(t *testing.T, db *sql.DB, email, password string) int64 {
	hashedPassword, err := utils.HashPassWord(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	result, err := db.Exec("INSERT INTO users(email, password) VALUES (?, ?)", email, hashedPassword)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userId, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	return userId
}
