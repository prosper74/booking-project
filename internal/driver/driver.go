package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB holds the database connection pool
type DB struct {
	SQL *sql.DB
}

var dbConnection = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

// ConnectSQL creates database pool for Postgres
func ConnectSQL(dbConnectionString string) (*DB, error) {
	newDatabase, err := NewDatabase(dbConnectionString)
	if err != nil {
		panic(err)
	}

	newDatabase.SetMaxOpenConns(maxOpenDbConn)
	newDatabase.SetMaxIdleConns(maxIdleDbConn)
	newDatabase.SetConnMaxLifetime(maxDbLifetime)

	dbConnection.SQL = newDatabase

	err = testDB(newDatabase)
	if err != nil {
		return nil, err
	}
	return dbConnection, nil
}

// testDB tries to ping the database
func testDB(newDatabase *sql.DB) error {
	err := newDatabase.Ping()
	if err != nil {
		return err
	}
	return nil
}

// NewDatabase creates a new database for the application
func NewDatabase(dbConnectionString string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dbConnectionString)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
