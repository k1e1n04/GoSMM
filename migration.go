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

// checkMigrationIntegrity checks the migration history table for inconsistencies
func checkMigrationIntegrity(db *sql.DB, config DBConfig) error {
	// Load executed migrations from the history table
	executedMigrations := make(map[string]bool)
	rows, err := db.Query(`SELECT filename FROM gosmm_migration_history WHERE success = TRUE`)
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

	// Check each executed migration exists in the migration directory
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".sql" {
			fmt.Errorf("Invalid file extension: %s", file.Name())
		}

		if executedMigrations[file.Name()] {
			delete(executedMigrations, file.Name())
		}
	}

	// Any remaining executed migrations in the map are inconsistencies
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

	// Get the last successful installed_rank
	lastInstalledRank, err := getLastInstalledRank(db)
	if err != nil {
		return err
	}

	// Check if there are any migrations where success is false
	failedMigrationExists, err := failedMigrationExists(db)
	if err != nil {
		return err
	}
	if failedMigrationExists {
		return fmt.Errorf("Cannot proceed, there is at least one failed migration")
	}

	// Get the filename of the last successful migration
	lastSuccessfulMigrationFile, err := getlastSuccessfulMigrationFile(db)
	if err != nil {
		return err
	}

	err = createHistoryTable(db)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
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

	installedRank := lastInstalledRank

	// If there are no executed migrations, then we should execute all migrations
	shouldExecute := lastSuccessfulMigrationFile == ""

	for _, file := range files {
		filename := file.Name()
		if executedMigrations[filename] {
			continue // skip already executed migrations
		}

		if shouldExecute {

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

				_, err = tx.Exec(statement)
				if err != nil {
					return fmt.Errorf("failed to execute statement: %s, error: %w", statement, err)
				}
			}
			executionTime := time.Since(startTime).Milliseconds()

			success := err == nil

			// Record the migration in the history table
			_, err = tx.Exec(`
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
		} else {
			if filename == lastSuccessfulMigrationFile {
				shouldExecute = true // Start executing from the next file
			}
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

// getLastInstalledRank returns the last successful installed_rank
func getLastInstalledRank(db *sql.DB) (int, error) {
	var lastInstalledRank int
	err := db.QueryRow("SELECT MAX(installed_rank) FROM gosmm_migration_history WHERE success = TRUE").Scan(&lastInstalledRank)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return lastInstalledRank, nil
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

// getlastSuccessfulMigrationFile returns the filename of the last successful migration
func getlastSuccessfulMigrationFile(db *sql.DB) (string, error) {
	var lastSuccessfulMigrationFile string
	err := db.QueryRow("SELECT filename FROM gosmm_migration_history WHERE success = TRUE ORDER BY installed_rank DESC LIMIT 1").Scan(&lastSuccessfulMigrationFile)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	return lastSuccessfulMigrationFile, nil
}
