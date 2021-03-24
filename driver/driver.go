package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbConnLifetime = 5 * time.Minute

// ConnectDB creates database pool for postgres and populates global dbConn.SQL
func ConnectDB(dsn string) (*DB, error) {
	db, err := CreateDB(dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(maxOpenDbConn)
	db.SetMaxIdleConns(maxIdleDbConn)
	db.SetConnMaxLifetime(maxDbConnLifetime)
	dbConn.SQL = db
	err = testDB(db)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// testDB tries to ping the given database
func testDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

// CreateDB creates a new database for the application
func CreateDB(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(); err != nil {
		return nil, err
	}

	return conn, nil
}
