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

func TestCheckMigrationIntegrity(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

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

	// Create gosmm_migration_history entry for the missing migration file
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

func TestCheckMigrationIntegrityWithInvalidExtension(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

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

func TestMigrateSingleFile(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

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

	err := Migrate(db, migrationsDir, "sqlite3")
	assert.NoError(t, err)

	// Check test_table exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'test_table')").Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check if test_table exists: %v", err)
	}
	if !exists {
		t.Fatalf("Failed to create test_table")
	}

	// Check gosmm_migration_history entry exists
	var installedRank int
	err = db.QueryRow("SELECT installed_rank FROM gosmm_migration_history WHERE filename = 'v20230101_create_test_data_00001.sql'").Scan(&installedRank)
	if err != nil {
		t.Fatalf("Failed to check if gosmm_migration_history entry exists: %v", err)
	}
	if installedRank != 1 {
		t.Fatalf("Failed to create gosmm_migration_history entry")
	}

	// Delete the test migration file
	if err := os.Remove(testMigrationFile); err != nil {
		t.Fatalf("Failed to delete test migration file: %v", err)
	}
}

func TestMigrateMultipleFiles(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create test_migrations directory if it doesn't exist
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		err := os.Mkdir(migrationsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test_migrations directory: %v", err)
		}
	}

	// Create test migration files in the test_migrations directory
	testMigrationFile1 := filepath.Join(migrationsDir, "v20230101_create_test_data_00001.sql")
	if err := ioutil.WriteFile(testMigrationFile1, []byte("CREATE TABLE test_table (id INTEGER);"), 0644); err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}
	testMigrationFile2 := filepath.Join(migrationsDir, "v20230101_create_test_data_00002.sql")
	if err := ioutil.WriteFile(testMigrationFile2, []byte("CREATE TABLE test_table_2 (id INTEGER);"), 0644); err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}

	err := Migrate(db, migrationsDir, "sqlite3")
	assert.NoError(t, err)

	// Check test_table exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'test_table')").Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check if test_table exists: %v", err)
	}
	if !exists {
		t.Fatalf("Failed to create test_table")
	}

	// Check test_table_2 exists
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'test_table_2')").Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check if test_table_2 exists: %v", err)
	}

	// Check gosmm_migration_history entry exists
	var installedRank int
	err = db.QueryRow("SELECT installed_rank FROM gosmm_migration_history WHERE filename = 'v20230101_create_test_data_00002.sql'").Scan(&installedRank)
	if err != nil {
		t.Fatalf("Failed to check if gosmm_migration_history entry exists: %v", err)
	}
	if installedRank != 2 {
		t.Fatalf("Failed to create gosmm_migration_history entry")
	}

	// Delete the test migration files
	if err := os.Remove(testMigrationFile1); err != nil {
		t.Fatalf("Failed to delete test migration file: %v", err)
	}
	if err := os.Remove(testMigrationFile2); err != nil {
		t.Fatalf("Failed to delete test migration file: %v", err)
	}
}

func TestMigrateWithSuccessFlagIsFalse(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

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

	// Create gosmm_migration_history entry
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

	err = Migrate(db, migrationsDir, "sqlite3")
	assert.Error(t, err)

	// Delete the test migration file
	if err := os.Remove(testMigrationFile); err != nil {
		t.Fatalf("Failed to delete test migration file: %v", err)
	}
}

func TestMigrateWithSecondFileHasInvalidSQL(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create test_migrations directory if it doesn't exist
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		err := os.Mkdir(migrationsDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test_migrations directory: %v", err)
		}
	}

	// Create test migration files in the test_migrations directory
	testMigrationFile1 := filepath.Join(migrationsDir, "v20230101_create_test_data_00001.sql")
	if err := ioutil.WriteFile(testMigrationFile1, []byte("CREATE TABLE test_table (id INTEGER);"), 0644); err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}
	testMigrationFile2 := filepath.Join(migrationsDir, "v20230101_create_test_data_00002.sql")
	if err := ioutil.WriteFile(testMigrationFile2, []byte(" TABLE test_table_2 (id INTEGER);"), 0644); err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}

	err := Migrate(db, migrationsDir, "sqlite3")
	assert.Error(t, err)

	// Check test_table exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'test_table')").Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check if test_table exists: %v", err)
	}
	if !exists {
		t.Fatalf("Failed to create test_table")
	}

	// // Check gosmm_migration_history entry exists
	var installedRank int
	var success bool
	err = db.QueryRow("SELECT installed_rank, success FROM gosmm_migration_history WHERE filename = 'v20230101_create_test_data_00001.sql'").Scan(&installedRank, &success)
	if err != nil {
		t.Fatalf("Failed to check if gosmm_migration_history entry exists: %v", err)
	}
	if installedRank != 1 {
		t.Fatalf("Failed to create gosmm_migration_history entry")
	}
	if success != true {
		t.Fatalf("Failed to create gosmm_migration_history entry")
	}

	err = db.QueryRow("SELECT installed_rank, success FROM gosmm_migration_history WHERE filename = 'v20230101_create_test_data_00002.sql'").Scan(&installedRank, &success)
	if err != nil {
		t.Fatalf("Failed to check if gosmm_migration_history entry exists: %v", err)
	}
	if installedRank != 2 {
		t.Fatalf("Failed to create gosmm_migration_history entry")
	}
	if success != false {
		t.Fatalf("Failed to create gosmm_migration_history entry")
	}

	// Delete the test migration files
	if err := os.Remove(testMigrationFile1); err != nil {
		t.Fatalf("Failed to delete test migration file: %v", err)
	}
	if err := os.Remove(testMigrationFile2); err != nil {
		t.Fatalf("Failed to delete test migration file: %v", err)
	}
}
