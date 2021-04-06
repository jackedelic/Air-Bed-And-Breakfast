package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/jackedelic/bookings/driver"
	"github.com/jackedelic/bookings/helpers"
	"github.com/jackedelic/bookings/internal/config"
	"github.com/jackedelic/bookings/internal/handlers"
	"github.com/jackedelic/bookings/internal/models"
	"github.com/jackedelic/bookings/internal/render"
	"github.com/jackedelic/bookings/repository/dbrepo"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer close(app.MailChan)

	listenForMail()

	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	handlers.Repo.DBRepo = dbrepo.NewPostgresRepo(db.SQL, &app)
	defer db.SQL.Close()

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	app.InfoLog.Println(fmt.Sprintf("Server listening at port %s", portNumber))
	err = srv.ListenAndServe()
	log.Fatal(err)
}

// run sets some app-wide configurations by populating global app
func run() error {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})

	// Initialize MailChan to app config
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.InProduction = false // Change this to true when in production
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Register session for all requests
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Error creating template cache", err)
		return err
	}

	app.TemplateCache = templateCache
	app.UseCache = false

	// handlers and render packages have access to the same config.AppConfig
	repo := handlers.NewRepo(&app, driver.DB{}) // create a new repo holding the app config we just created
	handlers.NewHandlers(repo)                  // assign this newly created repo to handlers.Repo
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return nil
}

// connectDB connects to the database using driver specified in driver package, and returns the *sql.DB
func connectDB() (*driver.DB, error) {
	// Connects to database
	var (
		host     = "127.0.0.1"
		port     = 5432
		database = "bookings"
		username = "postgres"
		password = "password"
	)

	conn, err := driver.ConnectDB(fmt.Sprintf("host=%s port=%d database=%s user=%s password=%s", host, port, database, username, password))
	app.InfoLog.Println("Connecting to our database...")

	if err != nil {
		app.ErrorLog.Panicln("Error connecting to our database :(")
		log.Fatal(err)
	}

	conn.SQL.Ping()
	app.InfoLog.Println("Successfully connected to database :)")

	return conn, nil
}
