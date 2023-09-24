package main

import (
	"fmt"
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

	port, err := strconv.Atoi(os.Getenv("GOSMM_PORT"))
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	config := gosmm.DBConfig{
		Driver:   os.Getenv("GOSMM_DRIVER"),
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

	switch command {
	case "status":
		if err := gosmm.DisplayStatus(db); err != nil {
			log.Fatalf("Restore failed: %v", err)
		}

	case "migrate":
		// Get migrations directory from environment variable
		migrationsDir := os.Getenv("GOSMM_MIGRATIONS_DIR")
		if migrationsDir == "" {
			migrationsDir = defaultMigrationsDir
		}
		// Perform database migration
		if err := gosmm.Migrate(db, migrationsDir); err != nil {
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
}
