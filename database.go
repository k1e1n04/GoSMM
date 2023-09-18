package gosmm

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DBConfig holds the database configuration information
type DBConfig struct {
	Driver        string
	Host          string
	Port          int
	User          string
	Password      string
	DBName        string
	MigrationsDir string
}

// LoadConfig loads the database configuration from a toml file
func LoadConfig(configFilePath string) (*DBConfig, error) {
	var config DBConfig
	if _, err := toml.DecodeFile(configFilePath, &config); err != nil {
		log.Fatalf("Could not read config: %v", err)
	}
	return &config, nil
}

// ConnectDB connects to the database based on the given DBConfig
func ConnectDB(config DBConfig) (*sql.DB, error) {
	var dsn string
	switch config.Driver {
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Password, config.DBName)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			config.User, config.Password, config.Host, config.Port, config.DBName)
	case "sqlite3":
		dsn = config.DBName // for SQLite, DBName is the file path
	default:
		return nil, fmt.Errorf("unsupported driver: %s", config.Driver)
	}

	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
