package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/atuprosper/booking-project/internal/config"
	"github.com/atuprosper/booking-project/internal/handlers"
	"github.com/atuprosper/booking-project/internal/models"
	"github.com/atuprosper/booking-project/internal/render"
)

const port = ":8080"

var app config.AppConfig
var session *scs.SessionManager

func main() {

	err := run()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("Server started at port %s", port))
	// Create a variable to serve the routes
	srv := &http.Server{
		Addr:    port,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() error {
	app.InProduction = false

	// Things to be stored in the session
	// gob, is a built in library used for stroing sessions
	gob.Register(models.Reservation{})

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
		return err
	}

	app.TemplateCache = tc
	app.UseCache = false

	// Variable to reference our app
	repo := handlers.NewRepo(&app)

	// Pass the repo variable back to the new handler
	handlers.NewHandlers(repo)

	// Render the NewTemplates and add a reference to the AppConfig
	render.NewTemplates(&app)

	return nil
}
