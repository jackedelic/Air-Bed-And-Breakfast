package dbrepo

import (
	"database/sql"

	"github.com/jackedelic/bookings/internal/config"
	"github.com/jackedelic/bookings/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig // why do we need *congif.AppConfig
	DB  *sql.DB
}

type testingDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewPostgresRepo initializes a postgresDBRepo (holding app config and a connected database)
// and returns it
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	repo := &postgresDBRepo{
		App: a,
		DB:  conn,
	}
	return repo
}

// NewTestingRepo initializes a testingDBRepo with the given AppConfig obj.
// It injects a dummy sql.DB database. We don't want to hit the database for unit testing.
func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testingDBRepo{
		App: a,
	}
}
