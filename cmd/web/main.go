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
	"github.com/joho/godotenv"
)

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

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

	fmt.Printf("Server started at host %s and port %s", host, port)
	// Create a variable to serve the routes
	srv := &http.Server{
		Addr:    host + ":" + port,
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
	gob.Register(models.TodoList{})
	gob.Register(make(map[string]int))

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	dbURI := os.Getenv("DBURI")

	// Read flags
	inProduction := flag.Bool("production", true, "App is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	// dbHost := flag.String("dbhost", "", "Database host")
	// dbName := flag.String("dbname", "", "Database name")
	// dbUser := flag.String("dbuser", "", "Database user")
	// dbPassword := flag.String("dbpassword", "", "Database password")
	// dbPort := flag.String("dbport", "", "Database port")
	// dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()

	// if *dbName == "" || *dbUser == "" || *dbHost == "" {
	// 	fmt.Println("Missing flag dependencies, attach the flag dependencies in your batch file")
	// 	os.Exit(1)
	// }

	if dbURI == "" {
		fmt.Println("Missing flag dependencies, attach the flag dependencies in your batch file")
		os.Exit(1)
	}

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
	// connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPassword, *dbSSL)
	connectedDB, err := driver.ConnectSQL(dbURI)
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
