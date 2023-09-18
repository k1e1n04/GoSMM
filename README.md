# GoSMM
Golang Simple Migration Manager

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

## Configuration
`GoSMM` uses a TOML configuration file to set up the database connection and other settings.

Here is an example `gosmm.toml` file:

    ```toml
    Driver = "mysql"
    Host = "localhost"
    Port = 3306
    User = "dbuser"
    Password = "dbpassword"
    DBName = "mydatabase"
    MigrationsDir = "./migrations"
    ```

You can load this configuration using the `LoadConfig()` function:

    ```go
    config, err := gosmm.LoadConfig("path/to/config.toml")
    ```

### Fields:
- `Driver`: Database driver ("postgres", "mysql", or "sqlite3").
- `Host`: Hostname of your database.
- `Port`: Port number for the database.
- `User`: Username for the database.
- `Password`: Password for the database.
- `DBName`: The name of the database.
- `MigrationsDir`: Directory where your SQL migration files are stored.

## Usage
### Performing Migrations
To perform migrations, use the Migrate function:

    ```go
    err := gosmm.Migrate(config)
    if err != nil {
    log.Fatalf("Migration failed: %v", err)
    }
    ```

This will:

1. Validate the DB configuration.
2. Connect to the database.
3. Check migration integrity.
4. Execute pending migrations.

## How to Contribute
Contributions are welcome! Feel free to submit a pull request on [GitHub](https://github.com/k1e1n04/gosmm).

## License
This project is licensed under the MIT License.

## Disclaimer
Please make sure to back up your database before running migrations. The maintainers are not responsible for any data loss.