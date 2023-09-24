package gosmm

import (
	"database/sql"
	"fmt"
)

// Restore cleans up the migration history by deleting records with success = false
func Restore(db *sql.DB) error {
	// Create history table if it doesn't exist
	err := createHistoryTable(db)

	// SQL query to delete records where success is false from the migration history table
	query := "DELETE FROM gosmm_migration_history WHERE success = FALSE"

	// Execute the DELETE query
	result, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute restore query: %w", err)
	}

	// Check how many rows were deleted
	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve the number of deleted rows: %w", err)
	}

	if rowsDeleted == 0 {
		fmt.Println("No records to restore.")
	} else {
		fmt.Printf("%d record(s) restored.\n", rowsDeleted)
	}

	return nil
}
