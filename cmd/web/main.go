package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/jackedelic/bookings/pkg/config"
	"github.com/jackedelic/bookings/pkg/handlers"
	"github.com/jackedelic/bookings/pkg/render"
)

const portNumber = ":8080"

var app config.AppConfig

var session *scs.SessionManager

func main() {
	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Error creating template cache")
	}

	app.TemplateCache = templateCache
	app.UseCache = false

	// handlers and render packages have access to the same config.AppConfig
	repo := handlers.NewRepo(&app) // create a new repo holding the app config we just created
	handlers.NewHandlers(repo)     // assign this newly created repo to handlers.Repo
	render.NewConfig(&app)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}
