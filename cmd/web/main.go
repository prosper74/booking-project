package main

import (
	"encoding/gob"
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

	// Close database connection when main function finish running
	defer connectedDB.SQL.Close()

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
	app.InProduction = false

	// Things to be stored in the session
	// gob, is a built in library used for stroing sessions
	gob.Register(models.Reservation{})

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
	connectedDB, err := driver.ConnectSQL("host=localhost port=5432 dbname=database_connect user=postgres password=")
	if err != nil {
		log.Fatal("Cannot connect to database. Closing application")
	}

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	// Variable to reference our app
	repo := handlers.NewRepo(&app, connectedDB)

	// Pass the repo variable back to the new handler
	handlers.NewHandlers(repo)

	// Render the NewTemplates and add a reference to the AppConfig
	render.NewTemplates(&app)

	// Pass the app config to the helpers
	helpers.NewHelpers(&app)

	return connectedDB, nil
}
