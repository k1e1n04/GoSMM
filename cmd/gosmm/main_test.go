package main

import (
	"bytes"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

const (
	migrationsDir = "./test_migrations"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func TestExecuteStatusCommand(t *testing.T) {
	// Set up a mock DB and teardown function
	db, teardown := setupTestDB(t)
	defer teardown()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test the "status" command
	err := executeCommand(db, "status")
	assert.NoError(t, err)

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Validate the output
	if !strings.Contains(output, "Migration Status:") {
		t.Errorf("Unexpected output: %s", output)
	}
}

func TestExecuteRestoreCommand(t *testing.T) {
	// Set up a mock DB and teardown function
	db, teardown := setupTestDB(t)
	defer teardown()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test the "restore" command
	err := executeCommand(db, "restore")
	assert.NoError(t, err)

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Validate the output
	if !strings.Contains(output, "No records to restore.") {
		t.Errorf("Unexpected output: %s", output)
	}
}

func TestExecuteMigrateCommand(t *testing.T) {
	// set GOSMM_MIGRATIONS_DIR to the test migrations directory
	os.Setenv("GOSMM_MIGRATIONS_DIR", migrationsDir)

	// Create test_migrations directory if it doesn't exist
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		err := os.Mkdir(migrationsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test_migrations directory: %v", err)
		}
	}

	// Set up a mock DB and teardown function
	db, teardown := setupTestDB(t)
	defer teardown()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test the "migrate" command
	err := executeCommand(db, "migrate")
	assert.NoError(t, err)

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Validate the output
	if !strings.Contains(output, "Migration completed successfully.") {
		t.Errorf("Unexpected output: %s", output)
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	// Set up a mock DB and teardown function
	db, teardown := setupTestDB(t)
	defer teardown()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test an unknown command
	err := executeCommand(db, "unknown")
	assert.NoError(t, err)

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Validate the output
	if !strings.Contains(output, "Unknown command: unknown") {
		t.Errorf("Unexpected output: %s", output)
	}
}
