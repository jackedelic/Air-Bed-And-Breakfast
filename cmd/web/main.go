package main

import (
	"encoding/gob"
	"flag"
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
	gob.Register(map[string]int{})

	// read flags
	inProd := flag.Bool("production", false, "Application is not in production mode by default")
	useCache := flag.Bool("cache", false, "Not using template cache by default")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbPort := flag.Int("dbport", 5432, "Database port is 5432 by default")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database username")
	dbPass := flag.String("dbpassword", "password", "Database password is password by default")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()

	app.InProduction = *inProd
	app.UseCache = *useCache
	app.DBHost = *dbHost
	app.DBPort = *dbPort
	app.DBName = *dbName
	app.DBUser = *dbUser
	app.DBPassword = *dbPass
	app.DBSSL = *dbSSL

	if app.DBName == "" || app.DBUser == "" {
		log.Println(app)
		log.Println("Misssing required flags")
		os.Exit(1)
	}

	// Initialize MailChan to app config
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

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
		host     = app.DBHost
		port     = app.DBPort
		database = app.DBName
		username = app.DBUser
		password = app.DBPassword
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
