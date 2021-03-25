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

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	repo := &postgresDBRepo{
		App: a,
		DB:  conn,
	}
	return repo
}
