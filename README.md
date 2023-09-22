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

Note: Please write your migration files in SQL format.

By emphasizing SQL-based migration files, GoSMM aims to provide a straightforward and consistent approach to database migration tasks.

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
    }
    ```

### Fields:
- `Driver`: Database driver ("postgres", "mysql", or "sqlite3").
- `Host`: Hostname of your database.
- `Port`: Port number for the database.
- `User`: Username for the database.
- `Password`: Password for the database.
- `DBName`: The name of the database.

### Performing Migrations
To perform migrations, use the Migrate function:

    ```go
    db, err := gosmm.ConnectDB(config)
    if err != nil {
        log.Fatalf("Connection failed: %v", err)
    }
    err := gosmm.Migrate(db, os.Getenv("MIGRATIONS_DIR"))
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    err := gosmm.CloseDB(db)
    if err != nil {
        log.Fatalf("Connection failed: %v", err)
    }
    ```

This will:

1. Connect to the database.
2. Validate the DB configuration.
3. Check migration integrity.
4. Execute pending migrations.
5. Record migration history.
6. Close the database connection.

## Migration History Table
`GoSMM` will create a migration history table in your database to keep track of which migrations have been executed. The table will be named `gosmm_migration_history` and will have the following schema:

| Column Name    | Data Type | Description                                     |
|----------------|-----------|-------------------------------------------------|
| installed_rank | int       | The rank of the migration.                      |
| filename       | TEXT      | The name of the migration script.               |
| installed_on   | TIMESTAMP | The timestamp when the migration was installed. |
| execution_time | int       | The time it took to execute the migration.      |
| success        | BOOLEAN   | Whether the migration was successful or not.    |

## How to Contribute
Contributions are welcome! Feel free to submit a pull request on [GitHub](https://github.com/k1e1n04/gosmm).

## License
This project is licensed under the MIT License.

## Disclaimer
Please make sure to back up your database before running migrations. The maintainers are not responsible for any data loss.