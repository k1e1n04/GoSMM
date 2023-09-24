package gosmm

import (
	"database/sql"
	"fmt"
)

// DisplayStatus displays the migration status
func DisplayStatus(db *sql.DB) error {
	// Create history table if it doesn't exist
	err := createHistoryTable(db)

	// SQL query to fetch migration statuses from the migration history table
	rows, err := db.Query("SELECT installed_rank, filename, installed_on, execution_time, success FROM gosmm_migration_history ORDER BY installed_rank ASC")
	if err != nil {
		fmt.Printf("Error fetching migration status: %v\n", err)
		return fmt.Errorf("failed to execute status query: %w", err)
	}
	defer rows.Close()

	fmt.Println("Migration Status:")
	fmt.Println("Rank | Filename | Installed On | Execution Time (ms) | Success")

	// Loop through each row in the result set
	for rows.Next() {
		if err != nil {
			fmt.Printf("Error creating history table: %v\n", err)
			return fmt.Errorf("failed to create history table: %w", err)
		}

		var (
			installedRank int
			filename      string
			installedOn   string
			executionTime int
			success       bool
		)

		if err := rows.Scan(&installedRank, &filename, &installedOn, &executionTime, &success); err != nil {
			fmt.Printf("Error reading row: %v\n", err)
			return fmt.Errorf("failed to read row: %w", err)
		}

		// Format success as text
		successText := "No"
		if success {
			successText = "Yes"
		}

		// Print the migration status
		fmt.Printf("%d | %s | %s | %d | %s\n", installedRank, filename, installedOn, executionTime, successText)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		fmt.Printf("Error iterating rows: %v\n", err)
		return fmt.Errorf("failed to iterate rows: %w", err)
	}

	return nil
}
