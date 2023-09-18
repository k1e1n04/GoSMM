package gosmm

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// checkMigrationIntegrity checks the integrity of the migrations by comparing the migration history
// with the actual files in the migration directory
func checkMigrationIntegrity(db *sql.DB, config DBConfig) error {
	executedMigrations := make(map[string]bool)
	rows, err := db.Query(`SELECT filename FROM gosmm_migration_history`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return err
		}
		executedMigrations[filename] = true
	}

	files, err := ioutil.ReadDir(config.MigrationsDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		if filepath.Ext(filename) != ".sql" {
			return fmt.Errorf("Inconsistent migration state. Non-SQL file found: %s", filename)
		}

		if !executedMigrations[filename] {
			return fmt.Errorf("Inconsistent migration state. Unexecuted migration file found: %s", filename)
		}
		delete(executedMigrations, filename)
	}

	for filename := range executedMigrations {
		return fmt.Errorf("Inconsistent migration state. Executed migration file not found: %s", filename)
	}

	return nil
}

// Migrate executes the SQL migrations in the given directory
func Migrate(config DBConfig) error {
	err := validateDBConfig(&config)
	if err != nil {
		return err
	}

	db, err := connectDB(config)
	if err != nil {
		return err
	}

	err = checkMigrationIntegrity(db, config)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = createHistoryTable(db)
	if err != nil {
		return err
	}

	// Load executed migrations from history table
	executedMigrations := make(map[string]bool)
	rows, err := db.Query(`SELECT filename FROM gosmm_migration_history`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return err
		}
		executedMigrations[filename] = true
	}

	// Read all SQL files from the migration directory
	files, err := ioutil.ReadDir(config.MigrationsDir)
	if err != nil {
		return err
	}

	// Sort files by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	installedRank := len(executedMigrations)

	for _, file := range files {
		filename := file.Name()
		if executedMigrations[filename] {
			continue // skip already executed migrations
		}

		installedRank++

		startTime := time.Now()

		// Read and execute the SQL file
		data, err := ioutil.ReadFile(filepath.Join(config.MigrationsDir, filename))
		if err != nil {
			return err
		}

		// Split the file content by ";" and execute each statement.
		statements := strings.Split(string(data), ";")
		for _, statement := range statements {
			statement = strings.TrimSpace(statement) // Trim whitespace
			if statement == "" {
				continue // Skip empty statements
			}

			_, err = db.Exec(statement)
			if err != nil {
				return fmt.Errorf("failed to execute statement: %s, error: %w", statement, err)
			}
		}
		executionTime := time.Since(startTime).Milliseconds()

		success := err == nil

		// Record the migration in the history table
		_, err = db.Exec(`
			INSERT INTO gosmm_migration_history (
				installed_rank, 
			 	filename, 
			    installed_on, 
			    execution_time, success
			) VALUES (?, ?, ?, ?, ?)
		`, installedRank, filename, startTime, executionTime, success)

		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("Migration failed for file: %s", filename)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
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
