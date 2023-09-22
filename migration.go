package gosmm

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	sqlFileExtension      = ".sql"
	migrationHistoryTable = "gosmm_migration_history"
)

// checkMigrationIntegrity checks the migration history table for inconsistencies
func checkMigrationIntegrity(db *sql.DB, migrationsDir string) error {
	// Load executed migrations from the history table
	executedMigrations := make(map[string]bool)
	rows, err := db.Query(`SELECT filename FROM gosmm_migration_history WHERE success = TRUE`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return err
		}
		executedMigrations[filename] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Read all SQL files from the migration directory
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	// Check each executed migration exists in the migration directory
	for _, file := range files {
		if filepath.Ext(file.Name()) != sqlFileExtension {
			return fmt.Errorf("invalid file extension: %s", file.Name())
		}

		if executedMigrations[file.Name()] {
			delete(executedMigrations, file.Name())
		}
	}

	// Any remaining executed migrations in the map are inconsistencies
	for filename := range executedMigrations {
		return fmt.Errorf("inconsistent migration state. executed migration file not found: %s", filename)
	}

	return nil
}

// Migrate executes the SQL migrations in the given directory
func Migrate(db *sql.DB, migrationsDir string) error {
	if err := createHistoryTable(db); err != nil {
		return fmt.Errorf("failed to create history table: %w", err)
	}

	if err := checkMigrationIntegrity(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to check migration integrity: %w", err)
	}

	lastInstalledRank, err := getLastInstalledRank(db)
	if err != nil {
		return fmt.Errorf("failed to get last successful installed_rank: %w", err)
	}

	failedMigrationExists, err := failedMigrationExists(db)
	if err != nil {
		return fmt.Errorf("failed to check if failed migration exists: %w", err)
	}
	if failedMigrationExists {
		return fmt.Errorf("cannot proceed, there is at least one failed migration")
	}

	lastSuccessfulMigrationFile, err := getLastSuccessfulMigrationFile(db)
	if err != nil {
		return err
	}

	executedMigrations, err := getExecutedMigrations(db)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	installedRank := lastInstalledRank
	shouldExecute := lastSuccessfulMigrationFile == ""

	for _, file := range files {
		filename := file.Name()

		if executedMigrations[filename] {
			continue // skip already executed migrations
		}

		if !shouldExecute && filename > lastSuccessfulMigrationFile {
			shouldExecute = true
		}

		if shouldExecute {
			installedRank++

			data, err := ioutil.ReadFile(filepath.Join(migrationsDir, filename))
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			statements := strings.Split(string(data), ";")

			if err := executeAndRecordMigration(db, tx, installedRank, filename, statements); err != nil {
				return err
			}
		}
	}

	return nil
}

// getExecutedMigrations returns a map of executed migrations
func getExecutedMigrations(db *sql.DB) (map[string]bool, error) {
	executedMigrations := make(map[string]bool)
	rows, err := db.Query(`SELECT filename FROM ` + migrationHistoryTable)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		executedMigrations[filename] = true
	}
	return executedMigrations, rows.Err()
}

// getLastInstalledRank returns the last successful installed_rank
func getLastSuccessfulMigrationFile(db *sql.DB) (string, error) {
	var lastSuccessfulMigrationFile string
	err := db.QueryRow(`SELECT filename FROM ` + migrationHistoryTable + ` WHERE success = TRUE ORDER BY installed_rank DESC LIMIT 1`).Scan(&lastSuccessfulMigrationFile)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return lastSuccessfulMigrationFile, nil
}

// executeAndRecordMigration executes the migration and records it in the history table
func executeAndRecordMigration(db *sql.DB, tx *sql.Tx, installedRank int, filename string, statements []string) error {
	startTime := time.Now()
	var success bool

	for _, statement := range statements {
		statement = strings.TrimSpace(statement) // Trim whitespace
		if statement == "" {
			continue // Skip empty statements
		}

		_, err := tx.Exec(statement)
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return fmt.Errorf("failed to rollback transaction error: %w original error: %w", e, err)
			}

			success = false
			tx, e = db.Begin()
			if e != nil {
				return fmt.Errorf("failed to begin error record transaction error: %w original error: %w", e, err)
			}
			e = recordMigration(tx, installedRank, filename, startTime, success)
			if e != nil {
				return fmt.Errorf("failed to record migration error: %w original error: %w", e, err)
			}
			return fmt.Errorf("failed to execute filename: %s, statement: %s, error: %w", filename, statement, err)
		}
	}

	success = true
	return recordMigration(tx, installedRank, filename, startTime, success)
}

// recordMigration records the migration in the history table
func recordMigration(tx *sql.Tx, installedRank int, filename string, startTime time.Time, success bool) error {
	executionTime := time.Since(startTime).Milliseconds()
	_, err := tx.Exec(`
		INSERT INTO `+migrationHistoryTable+` (
			installed_rank, 
			filename, 
			installed_on, 
			execution_time, 
			success
		) VALUES (?, ?, ?, ?, ?)
	`, installedRank, filename, startTime, executionTime, success)
	if err != nil {
		return fmt.Errorf("failed to record migration in history table, error: %w, filename: %s", err, filename)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// createHistoryTable creates the migration history table if it doesn't exist
func createHistoryTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS gosmm_migration_history (
		installed_rank INTEGER,
		filename TEXT,
		installed_on TIMESTAMP,
		execution_time INTEGER,
		success BOOLEAN
	)`)
	if err != nil {
		return err
	}
	return nil
}

// getLastInstalledRank returns the last successful installed_rank
func getLastInstalledRank(db *sql.DB) (int, error) {
	var lastInstalledRank sql.NullInt64
	err := db.QueryRow("SELECT MAX(installed_rank) FROM gosmm_migration_history WHERE success = TRUE").Scan(&lastInstalledRank)
	if err != nil {
		return 0, err
	}

	if !lastInstalledRank.Valid {
		return 0, nil
	}

	return int(lastInstalledRank.Int64), nil
}

// failedMigrationExists returns true if there is at least one failed migration
func failedMigrationExists(db *sql.DB) (bool, error) {
	var failedMigrationExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM gosmm_migration_history WHERE success = FALSE)").Scan(&failedMigrationExists)
	if err != nil {
		return false, err
	}
	return failedMigrationExists, nil
}
