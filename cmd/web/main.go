package main

import (
	"log"
	"net/http"

	"github.com/jackedelic/go-overview-trevor-sawler/pkg/config"
	"github.com/jackedelic/go-overview-trevor-sawler/pkg/handlers"
	"github.com/jackedelic/go-overview-trevor-sawler/pkg/render"
)

const portNumber = ":8080"

func main() {
	var app config.AppConfig
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Error creating template cache")
	}

	app.TemplateCache = templateCache
	app.UseCache = false
	render.NewConfig(&app)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}
