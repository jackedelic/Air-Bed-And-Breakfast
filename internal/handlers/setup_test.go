package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackedelic/bookings/internal/config"
	"github.com/jackedelic/bookings/internal/models"
	"github.com/jackedelic/bookings/internal/render"
	"github.com/justinas/nosurf"
)

var functions = template.FuncMap{}
var pathToTemplates = "../../templates"
var session *scs.SessionManager
var app config.AppConfig

func getRoutes() http.Handler {
	gob.Register(models.Reservation{})
	// Register session for all requests
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	templateCache, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("Error creating template cache")
	}

	app.TemplateCache = templateCache
	app.UseCache = true // if false, a route handler will call render.RenderTemplate which calls CreateTemplate,
	// but we don want to touch CreateTemplate (from render)

	// handlers and render packages have access to the same config.AppConfig
	repo := NewRepo(&app) // create a new repo holding the app config we just created
	NewHandlers(repo)     // assign this newly created repo to Repo
	render.NewConfig(&app)

	mux := chi.NewMux()
	mux.Use(middleware.Recoverer)
	// mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)
	mux.Get("/generals-quarters", Repo.GeneralsQuarters)
	mux.Get("/majors-suite", Repo.MajorsSuite)
	mux.Get("/contact", Repo.Contact)

	mux.Get("/search-availability", Repo.SearchAvailability)
	mux.Post("/search-availability", Repo.PostSearchAvailability)
	mux.Post("/receive-json", Repo.ReceiveJSON)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}

// NoSurf is an http middleware that attaches cookie to "/"
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads the session from app.AppConfig
func SessionLoad(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}

// CreateTestTemplateCache creates a mapping of template file name to its parsed template.
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	fmt.Println("create template cache")
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page) // last element of the path. e.g about.page.tmpl

		t, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// finds *.layout.tmpl files, just to check if there exists one.
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		// make sure there exists a *.layout.tmpl before we actually parse the template t
		if len(matches) > 0 {
			// parse this particular smtg.page.tmpl against all layout.tmpl files
			t, err = t.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates)) // same as t.ParseFiles("a.layout.tmpl","b.layout.tmpl"...)
			if err != nil {
				return myCache, err
			}
		}
		myCache[name] = t
	}
	return myCache, nil
}
