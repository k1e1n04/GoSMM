# GoSMM
Golang Simple Migration Manager

[![GoDoc](https://pkg.go.dev/badge/github.com/yourusername/yourreponame)](https://pkg.go.dev/github.com/k1e1n04/gosmm)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/yourusername/yourreponame.svg)](https://github.com/k1e1n04/gosmm/releases/latest)

## Overview
`GoSMM` is a simple yet powerful SQL migration manager written in Go. This library enables you to manage SQL database migrations for multiple database drivers including Postgres, MySQL, and SQLite3.

The `GoSMM` package allows you to perform tasks like:

- Checking migration integrity.
- Executing migration files in order.
- Creating migration history tables.
- Rollback transactions on failure.

## Installation
To get started, you can install `GoSMM` using go get:
    
    ```bash
    go get github.com/k1e1n04/gosmm
    ```


## Usage
### Configuration
    ```go
    config := gosmm.DBConfig{
        Driver:        os.Getenv("DB_DRIVER"),
        Host:          os.Getenv("DB_HOST"),
        Port:          os.Getenv("DB_PORT"),
        User:          os.Getenv("DB_USER"),
        Password:      os.Getenv("DB_PASSWORD"),
        DBName:        os.Getenv("DB_NAME"),
        MigrationsDir: os.Getenv("MIGRATIONS_DIR"),
    }
    ```

### Fields:
- `Driver`: Database driver ("postgres", "mysql", or "sqlite3").
- `Host`: Hostname of your database.
- `Port`: Port number for the database.
- `User`: Username for the database.
- `Password`: Password for the database.
- `DBName`: The name of the database.
- `MigrationsDir`: Directory where your SQL migration files are stored.

### Performing Migrations
To perform migrations, use the Migrate function:

    ```go
    db, err := gosmm.Connect(config)
    if err != nil {
        log.Fatalf("Connection failed: %v", err)
    }
    err := gosmm.Migrate(db, config)
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    ```

This will:

1. Connect to the database.
2. Validate the DB configuration.
3. Check migration integrity.
4. Execute pending migrations.

## How to Contribute
Contributions are welcome! Feel free to submit a pull request on [GitHub](https://github.com/k1e1n04/gosmm).

## License
This project is licensed under the MIT License.

## Disclaimer
Please make sure to back up your database before running migrations. The maintainers are not responsible for any data loss.