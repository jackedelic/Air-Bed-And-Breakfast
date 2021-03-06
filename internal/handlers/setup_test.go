package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackedelic/bookings/internal/config"
	"github.com/jackedelic/bookings/internal/models"
	"github.com/jackedelic/bookings/internal/render"
	"github.com/justinas/nosurf"
)

var functions = template.FuncMap{
	"humanDate":  render.HumanDate,
	"formatDate": render.FormatDate,
	"iterate":    render.Iterate,
}
var pathToTemplates = "../../templates"
var session *scs.SessionManager
var app config.AppConfig

func TestMain(m *testing.M) {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	// setup loggers for app config
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Register session for all requests
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// Initialize MailChan to app config
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)
	listenForMail()

	templateCache, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("Error creating template cache")
	}

	app.TemplateCache = templateCache
	app.UseCache = true // if false, a route handler will call render.Template which calls CreateTemplate,
	// but we don want to touch CreateTemplate (from render)

	// handlers and render packages have access to the same config.AppConfig
	repo := NewTestingRepo(&app) // create a new testing repo holding the app config we just created
	NewHandlers(repo)            // assign this newly created repo to Repo (global in handlers package)
	render.NewRenderer(&app)
	os.Exit(m.Run())
}

func listenForMail() {
	go func() {
		for {
			<-app.MailChan
		}
	}()
}

func getRoutes() http.Handler {
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

	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostLogin)
	mux.Get("/user/logout", Repo.Logout)

	mux.Post("/search-availability-json", Repo.SearchAvailabilityJSON)
	mux.Post("/search-availability", Repo.PostSearchAvailability)
	mux.Post("/receive-json", Repo.ReceiveJSON)

	mux.Route("/admin", func(mux chi.Router) {
		// mux.Use(Auth)
		mux.Get("/dashboard", Repo.AdminDashboard)
		mux.Get("/reservations-new", Repo.AdminNewReservations)
		mux.Get("/reservations-all", Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", Repo.AdminPostReservationCalendar)
		mux.Get("/reservations/{src}/{id}", Repo.AdminShowReservation)

		mux.Post("/reservations/{src}/{id}", Repo.AdminUpdateReservation)
		mux.Get("/process-reservation/{src}/{id}", Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}", Repo.AdminDeleteReservation)
	})

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
