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
	"github.com/atuprosper/booking-project/internal/config"
	"github.com/atuprosper/booking-project/internal/driver"
	"github.com/atuprosper/booking-project/internal/handlers"
	"github.com/atuprosper/booking-project/internal/helpers"
	"github.com/atuprosper/booking-project/internal/models"
	"github.com/atuprosper/booking-project/internal/render"
)

const port = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {

	connectedDB, err := run()

	if err != nil {
		log.Fatal(err)
	}

	// Close database connection when main function finish running and the mail server if mail has finsished sending
	defer connectedDB.SQL.Close()
	defer close(app.MailChannel)

	// Listening for mail
	fmt.Println("Listening for mail...")
	listenForMail()

	fmt.Println(fmt.Sprintf("Server started at port %s", port))
	// Create a variable to serve the routes
	srv := &http.Server{
		Addr:    port,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {

	// Things to be stored in the session
	// gob, is a built in library used for storing sessions
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(make(map[string]int))

	// Read flags
	inProduction := flag.Bool("production", true, "App is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "bookings", "Database name")
	dbUser := flag.String("dbuser", "postgres", "Database user")
	dbPassword := flag.String("dbpassword", "brokaarea24", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	// connectionString := fmt.Sprintf("host=%s port=%s dbname=bookings user=postgres password=brokaarea24 sslmode=")

	flag.Parse()

	app.InProduction = *inProduction
	app.UseCache = *useCache

	mailChannel := make(chan models.MailData)
	app.MailChannel = mailChannel

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// Connect to database
	log.Println("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPassword, *dbSSL)
	connectedDB, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connect to database. Closing application")
	}

	log.Println("Connected to database")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc	

	// Variable to reference our app
	repo := handlers.NewRepo(&app, connectedDB)

	// Pass the repo variable back to the new handler
	handlers.NewHandlers(repo)

	// Render the NewTemplates and add a reference to the AppConfig
	render.NewRenderer(&app)

	// Pass the app config to the helpers
	helpers.NewHelpers(&app)

	return connectedDB, nil
}
