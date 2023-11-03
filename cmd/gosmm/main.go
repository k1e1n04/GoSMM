package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/k1e1n04/gosmm/pkg/gosmm"
	"log"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

const (
	defaultMigrationsDir = "./migrations"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gosmm <command>")
		os.Exit(1)
	}

	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: Could not load .env file. If this is a production environment, ensure environment variables are set appropriately.")
	}

	port, err := strconv.Atoi(os.Getenv("GOSMM_PORT"))
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	driver := os.Getenv("GOSMM_DRIVER")

	config := gosmm.DBConfig{
		Driver:   driver,
		Host:     os.Getenv("GOSMM_HOST"),
		Port:     port,
		User:     os.Getenv("GOSMM_USER"),
		Password: os.Getenv("GOSMM_PASSWORD"),
		DBName:   os.Getenv("GOSMM_DBNAME"),
	}

	db, err := gosmm.ConnectDB(config)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	command := os.Args[1]

	if err := executeCommand(db, command, driver); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

func executeCommand(db *sql.DB, command string, driver string) error {
	switch command {
	case "status":
		if err := gosmm.DisplayStatus(db); err != nil {
			log.Fatalf("Status check failed: %v", err)
		}

	case "migrate":
		// Get migrations directory from environment variable
		migrationsDir := os.Getenv("GOSMM_MIGRATIONS_DIR")
		if migrationsDir == "" {
			migrationsDir = defaultMigrationsDir
		}
		// Perform database migration
		if err := gosmm.Migrate(db, migrationsDir, driver); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migration completed successfully.")

	case "restore":
		if err := gosmm.Restore(db); err != nil {
			log.Fatalf("Restore failed: %v", err)
		}

	default:
		fmt.Println("Unknown command:", command)
	}

	return nil
}
