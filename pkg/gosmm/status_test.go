package gosmm

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDisplayStatusWithNoHistoryRecord(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create gosmm_migration_history table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS gosmm_migration_history (
				installed_rank INTEGER,
				filename TEXT,
				installed_on TIMESTAMP,
				execution_time INTEGER,
				success BOOLEAN
            )`)
	if err != nil {
		t.Fatalf("Failed to create gosmm_migration_history table: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = DisplayStatus(db)
	if err != nil {
		t.Fatalf("Failed to display status: %v", err)
	}

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

func TestDisplayStatusWithOneHistoryRecord(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create gosmm_migration_history table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS gosmm_migration_history (
				installed_rank INTEGER,
				filename TEXT,
				installed_on TIMESTAMP,
				execution_time INTEGER,
				success BOOLEAN
            )`)
	if err != nil {
		t.Fatalf("Failed to create gosmm_migration_history table: %v", err)
	}

	// Insert a record into gosmm_migration_history
	_, err = db.Exec(
		`INSERT INTO gosmm_migration_history (
				installed_rank,
				filename,
				installed_on,
				execution_time,
				success
			) VALUES (?, ?, ?, ?, ?)`, 1, "test.sql", "2021-01-01 00:00:00", 0, true,
	)
	if err != nil {
		t.Fatalf("Failed to insert record into gosmm_migration_history: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = DisplayStatus(db)
	if err != nil {
		t.Fatalf("Failed to display status: %v", err)
	}

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

	if !strings.Contains(output, "1 | test.sql | 2021-01-01T00:00:00Z | 0 | Yes") {
		t.Errorf("Unexpected output: %s", output)
	}
}

func TestDisplayStatusWithFailedMigrationFile(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create gosmm_migration_history table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS gosmm_migration_history (
				installed_rank INTEGER,
				filename TEXT,
				installed_on TIMESTAMP,
				execution_time INTEGER,
				success BOOLEAN
			)`)
	if err != nil {
		t.Fatalf("Failed to create gosmm_migration_history table: %v", err)
	}

	// Insert a record into gosmm_migration_history
	_, err = db.Exec(
		`INSERT INTO gosmm_migration_history (
				installed_rank,
				filename,
				installed_on,
				execution_time,
				success
			) VALUES (?, ?, ?, ?, ?)`, 1, "test.sql", "2021-01-01 00:00:00", 0, false,
	)
	if err != nil {
		t.Fatalf("Failed to insert record into gosmm_migration_history: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = DisplayStatus(db)
	if err != nil {
		t.Fatalf("Failed to display status: %v", err)
	}

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

	if !strings.Contains(output, "1 | test.sql | 2021-01-01T00:00:00Z | 0 | No") {
		t.Errorf("Unexpected output: %s", output)
	}
}

func TestDisplayStatusWithMultipleHistoryRecords(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create gosmm_migration_history table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS gosmm_migration_history (
				installed_rank INTEGER,
				filename TEXT,
				installed_on TIMESTAMP,
				execution_time INTEGER,
				success BOOLEAN
			)`)
	if err != nil {
		t.Fatalf("Failed to create gosmm_migration_history table: %v", err)
	}

	// Insert a record into gosmm_migration_history
	_, err = db.Exec(
		`INSERT INTO gosmm_migration_history (
				installed_rank,
				filename,
				installed_on,
				execution_time,
				success
			) VALUES (?, ?, ?, ?, ?)`, 1, "test1.sql", "2021-01-01 00:00:00", 0, true,
	)
	if err != nil {
		t.Fatalf("Failed to insert record into gosmm_migration_history: %v", err)
	}

	// Insert a record into gosmm_migration_history
	_, err = db.Exec(
		`INSERT INTO gosmm_migration_history (
				installed_rank,
				filename,
				installed_on,
				execution_time,
				success
			) VALUES (?, ?, ?, ?, ?)`, 2, "test2.sql", "2021-01-02 00:00:00", 0, false,
	)
	if err != nil {
		t.Fatalf("Failed to insert record into gosmm_migration_history: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = DisplayStatus(db)
	if err != nil {
		t.Fatalf("Failed to display status: %v", err)
	}

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

	if !strings.Contains(output, "1 | test1.sql | 2021-01-01T00:00:00Z | 0 | Yes") {
		t.Errorf("Unexpected output: %s", output)
	}

	if !strings.Contains(output, "2 | test2.sql | 2021-01-02T00:00:00Z | 0 | No") {
		t.Errorf("Unexpected output: %s", output)
	}
}
