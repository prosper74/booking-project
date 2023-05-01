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
	"github.com/atuprosper/booking-project/internal/config"
	"github.com/atuprosper/booking-project/internal/models"
	"github.com/atuprosper/booking-project/internal/render"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/justinas/nosurf"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates = "./../../templates"

var functions = template.FuncMap{
	"humanDate":  render.HumanDate,
	"formatDate": render.FormatDate,
	"iterate":    render.Iterate,
	"add":        render.Add,
}

func TestMain(m *testing.M) {
	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.TodoList{})
	gob.Register(map[string]int{})

	// change this to true when in production
	app.InProduction = false

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// set up the session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	mailChannel := make(chan models.MailData)
	app.MailChannel = mailChannel
	defer close(mailChannel)

	listenForMail()

	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = true

	repo := NewTestRepo(&app)
	NewHandlers(repo)

	render.NewRenderer(&app)

	os.Exit(m.Run())
}

func listenForMail() {
	go func() {
		for {
			_ = <-app.MailChannel
		}
	}()
}

func getRoutes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/contact", Repo.Contact)
	mux.Get("/rooms/{id}", Repo.SingleRoom)

	mux.Get("/reservation", Repo.Reservation)
	mux.Post("/reservation", Repo.PostReservation)
	mux.Post("/reservation-json", Repo.AvailabilityJSON)
	mux.Get("/choose-room/{id}", Repo.ChooseRoom)
	mux.Get("/book-room", Repo.BookRoom)

	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostMakeReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	mux.Get("/user/login", Repo.Login)
	mux.Post("/user/login", Repo.PostLogin)
	mux.Get("/user/logout", Repo.Logout)

	mux.Get("/dashboard", Repo.AdminDashboard)

	mux.Get("/new-reservations", Repo.AdminNewReservations)
	mux.Get("/all-reservations", Repo.AdminAllReservations)
	mux.Get("/reservations-calendar", Repo.AdminReservationsCalendar)
	mux.Post("/reservations-calendar", Repo.AdminPostReservationsCalendar)

	mux.Get("/reservations/{src}/{id}/show", Repo.AdminSingleReservation)
	mux.Post("/reservations/{src}/{id}", Repo.PostAdminSingleReservation)

	mux.Get("/admin/rooms", Repo.AdminAllRooms)
	mux.Get("/admin/rooms/{id}", Repo.AdminSingleRoom)
	mux.Post("/admin/rooms/{id}", Repo.PostAdminSingleRoom)
	mux.Get("/admin/rooms/new-room", Repo.AdminNewRoom)
	mux.Post("/admin/rooms/new-room", Repo.PostAdminNewRoom)
	mux.Get("/admin/delete-room/{id}", Repo.AdminDeleteRoom)

	mux.Get("/admin/process-reservation/{src}/{id}/do", Repo.AdminProcessReservation)
	mux.Get("/admin/delete-reservation/{src}/{id}/do", Repo.AdminDeleteReservation)

	mux.Get("/admin/todo-list", Repo.AdminTodoList)
	mux.Post("/admin/todo-list", Repo.PostAdminTodoList)
	mux.Get("/admin/delete-todo/{id}", Repo.AdminDeleteTodo)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}

// NoSurf is the csrf protection middleware
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

// SessionLoad loads and saves session data for current request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// CreateTestTemplateCache creates a template cache as a map
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.html", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
