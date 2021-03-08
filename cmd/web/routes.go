package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackedelic/go-overview-trevor-sawler/pkg/config"
	"github.com/jackedelic/go-overview-trevor-sawler/pkg/handlers"
)

func routes(app *config.AppConfig) http.Handler {
	// mux := pat.New()
	// mux.Get("/", http.HandlerFunc(handlers.Repo.Home))
	// mux.Get("/about", http.HandlerFunc(handlers.Repo.About))
	mux := chi.NewMux()
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	return mux
}
