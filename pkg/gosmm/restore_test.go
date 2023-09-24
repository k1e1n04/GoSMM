package gosmm

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestRestore(t *testing.T) {
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

	// Insert some records into gosmm_migration_history table
	_, err = db.Exec(`INSERT INTO gosmm_migration_history (installed_rank, filename, installed_on, execution_time, success) VALUES
		(1, 'file1.sql', '2021-01-01 12:34:56', 123, TRUE),
		(2, 'file2.sql', '2021-01-01 12:34:56', 123, TRUE),
		(3, 'file3.sql', '2021-01-01 12:34:56', 123, TRUE),
		(4, 'file4.sql', '2021-01-01 12:34:56', 123, FALSE)`)
	if err != nil {
		t.Fatalf("Failed to insert records: %v", err)
	}

	// Call the Restore function
	err = Restore(db)
	if err != nil {
		t.Fatalf("Restore function failed: %v", err)
	}

	// Validate that records with success=FALSE are deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM gosmm_migration_history WHERE success = FALSE").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 records with success=FALSE, but got %d", count)
	}
}
