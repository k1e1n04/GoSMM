package gosmm

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"time"
)

func Migrate(db *sql.DB, config DBConfig) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Create history table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS gosmm_migration_history (
		installed_rank INTEGER,
		filename TEXT,
		installed_on TIMESTAMP,
		execution_time INTEGER,
		success BOOLEAN
	)`)
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

	installedRank := 0

	for _, file := range files {
		filename := file.Name()
		if executedMigrations[filename] {
			continue // skip already executed migrations
		}
		if filepath.Ext(filename) != ".sql" {
			continue // skip non-SQL files
		}

		installedRank++

		startTime := time.Now()

		// Read and execute the SQL file
		data, err := ioutil.ReadFile(filepath.Join(config.MigrationsDir, filename))
		if err != nil {
			return err
		}

		_, err = db.Exec(string(data))
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

	return nil
}
