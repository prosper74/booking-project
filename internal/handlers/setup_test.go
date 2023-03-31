package handlers

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
var functions = template.FuncMap{}

func TestMain(m *testing.M) {
	// what am I going to put in the session
	gob.Register(models.Reservation{})

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

	// connectedDB, err := driver.ConnectSQL("host=localhost port=5432 dbname=test_database user=postgres password=")
	// if err != nil {
	// 	log.Fatal("Cannot connect to database. Closing application")
	// }

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

func getRoutes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/contact", Repo.Contact)
	mux.Get("/alpine", Repo.Alpine)
	mux.Get("/generals", Repo.Generals)

	mux.Get("/reservation", Repo.Reservation)
	mux.Post("/reservation", Repo.PostReservation)
	mux.Post("/reservation-json", Repo.AvailabilityJSON)

	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostMakeReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

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

func TestRepository_MakeReservation(t *testing.T) {
	reservation := models.Reservation{
		ID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Generals Suit",
		},
	}

	request, _ := http.NewRequest("GET", "/make-reservation", nil)
	requestContext := getContext(request)
	request = request.WithContext(requestContext)

	// NewRecorder assimilates a request response cycle like a browser
	responseRecorder := httptest.NewRecorder()

	session.Put(requestContext, "reservation", reservation)
	handler := http.HandlerFunc(Repo.MakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	// Check if test pass
	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusOK)
	}

	// Test cases when reservation is not in session (Reset everything)
	request, _ = http.NewRequest("GET", "/make-reservation", nil)
	requestContext = getContext(request)
	request = request.WithContext(requestContext)
	responseRecorder = httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code when reservation is not in session: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test with non-existing room
	reservation.RoomID = 152221
	session.Put(requestContext, "reservation", reservation)
	request, _ = http.NewRequest("GET", "/make-reservation", nil)
	requestContext = getContext(request)
	request = request.WithContext(requestContext)
	responseRecorder = httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code when reservation is not in session: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostMakeReservation(t *testing.T) {
	requestBody := "start_date=2050-01-01"

	postedData := url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-05")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "atu@prosper.com")
	postedData.Add("phone", "484848448484")
	postedData.Add("room_id", "1")

	request, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	requestContext := getContext(request)
	request = request.WithContext(requestContext)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for missing body
	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	requestContext = getContext(request)
	request = request.WithContext(requestContext)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid start date
	postedData.Add("start_date", "invalid")
	postedData.Add("end_date", "2050-01-05")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "atu@prosper.com")
	postedData.Add("phone", "484848448484")
	postedData.Add("room_id", "1")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	requestContext = getContext(request)
	request = request.WithContext(requestContext)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid start date: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid end date
	postedData.Add("start_date", "2050-01-02")
	postedData.Add("end_date", "invalid")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "atu@prosper.com")
	postedData.Add("phone", "484848448484")
	postedData.Add("room_id", "1")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	requestContext = getContext(request)
	request = request.WithContext(requestContext)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid end date: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid room id
	postedData.Add("start_date", "2050-01-02")
	postedData.Add("end_date", "2050-01-05")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "atu@prosper.com")
	postedData.Add("phone", "484848448484")
	postedData.Add("room_id", "invalid")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	requestContext = getContext(request)
	request = request.WithContext(requestContext)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid room id: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid data
	requestBody = "start_date=2050-01-02"
	requestBody = fmt.Sprintf("%s&%s", requestBody, "end_date=2050-01-05")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "first_name=P")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "last_name=Atu")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "email=atu@prosper.com")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "phone=145245254554")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "room_id=1")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	requestContext = getContext(request)
	request = request.WithContext(requestContext)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid data: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for inability to insert reservation
	requestBody = "start_date=2050-01-02"
	requestBody = fmt.Sprintf("%s&%s", requestBody, "end_date=2050-01-05")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "first_name=Prosper")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "last_name=Atu")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "email=atu@prosper.com")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "phone=145245254554")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "room_id=2")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	requestContext = getContext(request)
	request = request.WithContext(requestContext)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for failing to insert reservation: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	// First case - Rooms are not available
	requestBody := "start=2050-01-01"
	requestBody = fmt.Sprintf("%s&%s", requestBody, "end=2050-01-05")
	requestBody = fmt.Sprintf("%s&%s", requestBody, "room_id=1")

	request, _ := http.NewRequest("POST", "/reservation-json", strings.NewReader(requestBody))
	requestContext := getContext(request)
	request = request.WithContext(requestContext)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler := http.HandlerFunc(Repo.AvailabilityJSON)
	responseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	var j jsonResponse
	// convert the json response and save it in the variable j
	err := json.Unmarshal([]byte(responseRecorder.Body.String()), &j)
	if err != nil {
		t.Error("Failed to parse JSON")
	}
}

func getContext(request *http.Request) context.Context {
	context, err := session.Load(request.Context(), request.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return context
}
