package gosmm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
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

func TestCheckMigrationIntegrity(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	migrationsDir := "./test_migrations"

	// Create test_migrations directory if it doesn't exist
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		err := os.Mkdir(migrationsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test_migrations directory: %v", err)
		}
	}

	// Create a test migration file in the test_migrations directory
	testMigrationFile := filepath.Join(migrationsDir, "v20230101_create_test_data_00001.sql")
	if err := ioutil.WriteFile(testMigrationFile, []byte("CREATE TABLE test_table (id INTEGER);"), 0644); err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}

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

	err = checkMigrationIntegrity(db, migrationsDir)
	assert.NoError(t, err)

	// Delete the test migration file
	if err := os.Remove(testMigrationFile); err != nil {
		t.Fatalf("Failed to delete test migration file: %v", err)
	}
}

func TestCheckMigrationIntegrityWithMissingMigrationFile(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	migrationsDir := "./test_migrations"

	// Create test_migrations directory if it doesn't exist
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		err := os.Mkdir(migrationsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test_migrations directory: %v", err)
		}
	}

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

	// Create gossm_migration_history entry for the missing migration file
	_, err = db.Exec(`INSERT INTO gosmm_migration_history (
			installed_rank,
			filename,
			installed_on,
			execution_time,
			success
		) VALUES (?, ?, ?, ?, ?)`, 1, "v20230101_create_test_data_00001.sql", "2021-01-01 00:00:00", 0, 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert gosmm_migration_history entry: %v", err)
	}

	err = checkMigrationIntegrity(db, migrationsDir)
	assert.Error(t, err)
}

func TestCheckMigrationIntegrityWithInvalidExtentionFile(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	migrationsDir := "./test_migrations"

	// Create test_migrations directory if it doesn't exist
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		err := os.Mkdir(migrationsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test_migrations directory: %v", err)
		}
	}

	// Create a test migration file in the test_migrations directory
	testMigrationFile := filepath.Join(migrationsDir, "v20230101_create_test_data_00001.txt")
	if err := ioutil.WriteFile(testMigrationFile, []byte("CREATE TABLE test_table (id INTEGER);"), 0644); err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}

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

	err = checkMigrationIntegrity(db, migrationsDir)
	assert.Error(t, err)

	// Delete the test migration file
	if err := os.Remove(testMigrationFile); err != nil {
		t.Fatalf("Failed to delete test migration file: %v", err)
	}
}

func TestMigrateWithSuccessFlagIsFalse(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	migrationsDir := "./test_migrations"

	// Create test_migrations directory if it doesn't exist
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		err := os.Mkdir(migrationsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test_migrations directory: %v", err)
		}
	}

	// Create a test migration file in the test_migrations directory
	testMigrationFile := filepath.Join(migrationsDir, "v20230101_create_test_data_00001.sql")
	if err := ioutil.WriteFile(testMigrationFile, []byte("CREATE TABLE test_table (id INTEGER);"), 0644); err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}

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

	// Create gossm_migration_history entry
	_, err = db.Exec(`INSERT INTO gosmm_migration_history (
			installed_rank,
			filename,
			installed_on,
			execution_time,
			success
		) VALUES (?, ?, ?, ?, ?)`, 1, "v20230101_create_test_data_00001.sql", "2021-01-01 00:00:00", 0, 0,
	)
	if err != nil {
		t.Fatalf("Failed to insert gosmm_migration_history entry: %v", err)
	}

	err = Migrate(db, migrationsDir)
	assert.Error(t, err)

	// Delete the test migration file
	if err := os.Remove(testMigrationFile); err != nil {
		t.Fatalf("Failed to delete test migration file: %v", err)
	}
}
