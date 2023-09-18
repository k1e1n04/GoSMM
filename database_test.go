package gosmm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateDBConfig(t *testing.T) {
	validConfig := DBConfig{
		Driver:        "mysql",
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "password",
		DBName:        "test_db",
		MigrationsDir: "migrations",
	}

	assert.Nil(t, validateDBConfig(&validConfig))
}

func TestValidateDBConfigWithInvalidPort(t *testing.T) {
	validConfig := DBConfig{
		Driver:        "mysql",
		Host:          "localhost",
		Port:          65536,
		User:          "root",
		Password:      "password",
		DBName:        "test_db",
		MigrationsDir: "migrations",
	}

	assert.NotNil(t, validateDBConfig(&validConfig))
}

func TestValidateDBConfigWithMissingDriver(t *testing.T) {
	validConfig := DBConfig{
		Driver:        "",
		Host:          "localhost",
		Port:          65536,
		User:          "root",
		Password:      "password",
		DBName:        "test_db",
		MigrationsDir: "migrations",
	}

	assert.NotNil(t, validateDBConfig(&validConfig))
}

func TestValidateDBConfigWithMissingHost(t *testing.T) {
	validConfig := DBConfig{
		Driver:        "mysql",
		Host:          "",
		Port:          65536,
		User:          "root",
		Password:      "password",
		DBName:        "test_db",
		MigrationsDir: "migrations",
	}

	assert.NotNil(t, validateDBConfig(&validConfig))
}

func TestValidateDBConfigWithMissingUser(t *testing.T) {
	validConfig := DBConfig{
		Driver:        "mysql",
		Host:          "localhost",
		Port:          65536,
		User:          "",
		Password:      "password",
		DBName:        "test_db",
		MigrationsDir: "migrations",
	}

	assert.NotNil(t, validateDBConfig(&validConfig))
}

func TestValidateDBConfigWithMissingPassword(t *testing.T) {
	validConfig := DBConfig{
		Driver:        "mysql",
		Host:          "localhost",
		Port:          65536,
		User:          "root",
		Password:      "",
		DBName:        "test_db",
		MigrationsDir: "migrations",
	}

	assert.NotNil(t, validateDBConfig(&validConfig))
}

func TestValidateDBConfigWithMissingDBName(t *testing.T) {
	validConfig := DBConfig{
		Driver:        "mysql",
		Host:          "localhost",
		Port:          65536,
		User:          "root",
		Password:      "password",
		DBName:        "",
		MigrationsDir: "migrations",
	}

	assert.NotNil(t, validateDBConfig(&validConfig))
}

func TestValidateDBConfigWithMissingMigrationsDir(t *testing.T) {
	validConfig := DBConfig{
		Driver:        "mysql",
		Host:          "localhost",
		Port:          65536,
		User:          "root",
		Password:      "password",
		DBName:        "test_db",
		MigrationsDir: "",
	}

	assert.NotNil(t, validateDBConfig(&validConfig))
}

func TestConnectDBWithMySQL(t *testing.T) {
	config := DBConfig{
		Driver:        "mysql",
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "password",
		DBName:        "test_db",
		MigrationsDir: "migrations",
	}

	_, err := connectDB(config)
	assert.Nil(t, err)
}

func TestConnectDBWithPostgres(t *testing.T) {
	config := DBConfig{
		Driver:        "postgres",
		Host:          "localhost",
		Port:          5432,
		User:          "root",
		Password:      "password",
		DBName:        "test_db",
		MigrationsDir: "migrations",
	}

	_, err := connectDB(config)
	assert.Nil(t, err)
}

func TestConnectDBWithSQLite(t *testing.T) {
	config := DBConfig{
		Driver:        "sqlite3",
		Host:          "localhost",
		Port:          5432,
		User:          "root",
		Password:      "password",
		DBName:        "test_db",
		MigrationsDir: "migrations",
	}

	_, err := connectDB(config)
	assert.Nil(t, err)
}
