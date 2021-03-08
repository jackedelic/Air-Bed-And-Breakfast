package main

import (
	"log"
	"net/http"

	"github.com/jackedelic/go-overview-trevor-sawler/pkg/config"
	"github.com/jackedelic/go-overview-trevor-sawler/pkg/handlers"
	"github.com/jackedelic/go-overview-trevor-sawler/pkg/render"
)

func main() {
	var app config.AppConfig
	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Error creating template cache")
	}
	app.TemplateCache = templateCache
	render.NewConfig(&app)
	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/about", handlers.About)
	_ = http.ListenAndServe(":8080", nil)
}
