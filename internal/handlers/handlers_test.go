package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/atuprosper/booking-project/internal/driver"
	"github.com/atuprosper/booking-project/internal/models"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"room", "/r", "/rooms/1", http.StatusOK},
	{"reservation", "/reservation", "GET", http.StatusOK},
	{"make res", "/make-reservation", "GET", http.StatusOK},
	{"choose room", "/choose-room/1", "GET", http.StatusOK},
	{"book room", "/book-room", "GET", http.StatusOK},
	{"res summary", "/reservation-summary", "GET", http.StatusOK},
	{"non-existent", "/green/eggs/and/ham", "GET", http.StatusNotFound},
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"dasboard", "/admin/dasboard", "GET", http.StatusOK},
	{"new res", "/admin/new-reservations", "GET", http.StatusOK},
	{"all res", "/admin/all-reservations", "GET", http.StatusOK},
	{"res cal", "/admin/reservations-calendar", "GET", http.StatusOK},
	{"show res cal with params", "/admin/reservations-calendar?y=2020&m=1", "GET", http.StatusOK},
	{"single res", "/admin/reservations/new/1/show", "GET", http.StatusOK},
	{"admin rooms", "/admin/rooms", "GET", http.StatusOK},
	{"new room", "/admin/rooms/new-room", "GET", http.StatusOK},
	{"single room", "/admin/rooms/1", "GET", http.StatusOK},
	{"new room", "/admin/rooms/new-room", "GET", http.StatusOK},
	{"todo", "/admin/todo-list", "GET", http.StatusOK},
}

func TestHandlers(testPointer *testing.T) {
	routes := getRoutes()

	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close()

	for _, element := range theTests {
		response, err := testServer.Client().Get(testServer.URL + element.url)
		if err != nil {
			testPointer.Log(err)
			testPointer.Fatal(err)
		}

		if response.StatusCode != element.expectedStatusCode {
			testPointer.Errorf("for %s expected %d but got %d", element.name, element.expectedStatusCode, response.StatusCode)
		}
	}
}

// data for the Reservation handler, /make-reservation route
var reservationTests = []struct {
	name               string
	reservation        models.Reservation
	expectedStatusCode int
	expectedLocation   string
	expectedHTML       string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "Generals Suit",
			},
		},
		expectedStatusCode: http.StatusOK,
		expectedHTML:       `action="/make-reservation"`,
	},
	{
		name:               "reservation-not-in-session",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	{
		name: "non-existent-room",
		reservation: models.Reservation{
			RoomID: 100,
			Room: models.Room{
				ID:       100,
				RoomName: "Generals Suit",
			},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
}

func TestRepository_MakeReservation(t *testing.T) {
	for _, e := range reservationTests {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getContext(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.Reservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}

		if e.expectedLocation != "" {
			// get the URL from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}

// postReservationTests is the test data for hte PostReservation handler test
var postReservationTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
}{
	{
		name: "valid-data",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"Prosper"},
			"last_name":  {"Atu"},
			"email":      {"atu@prosper.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/reservation-summary",
	},
	{
		name:                 "missing-post-body",
		postedData:           nil,
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-start-date",
		postedData: url.Values{
			"start_date": {"invalid"},
			"end_date":   {"2050-01-02"},
			"first_name": {"Prosper"},
			"last_name":  {"Atu"},
			"email":      {"atu@prosper.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-end-date",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"end"},
			"first_name": {"Prosper"},
			"last_name":  {"Atu"},
			"email":      {"atu@prosper.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-room-id",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"Prosper"},
			"last_name":  {"Atu"},
			"email":      {"atu@prosper.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"invalid"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-data",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"P"},
			"last_name":  {"Atu"},
			"email":      {"atu@prosper.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusOK,
		expectedHTML:         `action="/make-reservation"`,
		expectedLocation:     "",
	},
	{
		name: "database-insert-fails-reservation",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"Prosper"},
			"last_name":  {"Atu"},
			"email":      {"atu@prosper.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"2"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "database-insert-fails-restriction",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1000"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
}

func TestRepository_PostMakeReservation(t *testing.T) {
	// create our request body
	postedData := url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "prosper@atu.com")
	postedData.Add("phone", "2255887744")
	postedData.Add("room_id", "1")

	// create our request
	request, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))

	// get the context with session
	ctx := getContext(request)
	request = request.WithContext(ctx)

	// set the request header
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// create our response recorder, which satisfies the requirements
	// for http.ResponseWriter
	responseRecorder := httptest.NewRecorder()

	// make our handler a http.HandlerFunc
	// make the request to our handler
	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// test for missing post body
	request, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid start date
	postedData = url.Values{}
	postedData.Add("start_date", "invalid")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "prosper@atu.com")
	postedData.Add("phone", "2255887744")
	postedData.Add("room_id", "1")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid start date: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid end date
	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "invalid")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "prosper@atu.com")
	postedData.Add("phone", "2255887744")
	postedData.Add("room_id", "1")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid end date: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid room id
	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "prosper@atu.com")
	postedData.Add("phone", "2255887744")
	postedData.Add("room_id", "invalid")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid room id: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// test for failure to insert reservation into database
	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "prosper@atu.com")
	postedData.Add("phone", "2255887744")
	postedData.Add("room_id", "2")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed when trying to insert reservation: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// test for failure to insert restriction into database
	postedData = url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "prosper@atu.com")
	postedData.Add("phone", "2255887744")
	postedData.Add("room_id", "10000")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed when trying to insert restrictions: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	// first case -- rooms are not available
	postedData := url.Values{}
	postedData.Add("start", "2050-01-01")
	postedData.Add("end", "2050-01-02")
	postedData.Add("room_id", "1")

	request, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx := getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(responseRecorder, request)

	// since we have no rooms available, we expect to get status http.StatusSeeOther
	// this time we want to parse JSON and get the expected response
	var j jsonResponse
	err := json.Unmarshal([]byte(responseRecorder.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// since we specified a start date > 2049-12-31, we expect no availability
	if j.Ok {
		t.Error("Got availability when none was expected in AvailabilityJSON")
	}

	// second case -- rooms not available
	postedData = url.Values{}
	postedData.Add("start", "2040-01-01")
	postedData.Add("end", "2040-01-02")
	postedData.Add("room_id", "1")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(responseRecorder, request)

	// this time we want to parse JSON and get the expected response
	err = json.Unmarshal([]byte(responseRecorder.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// since we specified a start date < 2049-12-31, we expect availability
	if !j.Ok {
		t.Error("Got no availability when some was expected in AvailabilityJSON")
	}

	// third case -- no request body
	request, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(responseRecorder, request)

	// this time we want to parse JSON and get the expected response
	err = json.Unmarshal([]byte(responseRecorder.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// since we specified a start date < 2049-12-31, we expect availability
	if j.Ok || j.Message != "Internal server error" {
		t.Error("Got availability when request body was empty")
	}

	// fourth case -- database error
	postedData = url.Values{}
	postedData.Add("start", "2040-01-01")
	postedData.Add("end", "2040-01-02")
	postedData.Add("room_id", "1")

	request, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(responseRecorder, request)

	// this time we want to parse JSON and get the expected response
	err = json.Unmarshal([]byte(responseRecorder.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// since we specified a start date < 2049-12-31, we expect availability
	if j.Ok || j.Message != "Error querying database" {
		t.Error("Got availability when simulating database error")
	}
}

func TestRepository_ReservationSummary(t *testing.T) {
	// first case -- reservation in session
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Generals Quarters",
		},
	}

	request, _ := http.NewRequest("POST", "/make-reservation", nil)
	ctx := getContext(request)
	request = request.WithContext(ctx)
	responseRecorder := httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusOK)
	}

	// second case -- reservation not in session
	request, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getContext(request)
	request = request.WithContext(ctx)
	responseRecorder = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusOK)
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	// first case -- reservation in session
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Generals Suit",
		},
	}

	request, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getContext(request)
	request = request.WithContext(ctx)
	request.RequestURI = "/choose-room/1"

	responseRecorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// second case -- reservation not in session
	request, _ = http.NewRequest("GET", "/choose-room/1", nil)
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.RequestURI = "/choose-room/1"

	responseRecorder = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusSeeOther)
	}

	// third case -- missing url parameter, or malformed parameter
	request, _ = http.NewRequest("GET", "/choose-room/goat", nil)
	ctx = getContext(request)
	request = request.WithContext(ctx)
	request.RequestURI = "/choose-room/goat"

	responseRecorder = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code for missing parameter: got %d, wanted %d", responseRecorder.Code, http.StatusSeeOther)
	}
}

func TestRepository_BookRoom(t *testing.T) {
	// first case -- reservation in session
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Generals Suit",
		},
	}

	request, _ := http.NewRequest("GET", "/book-room?sd=2050-01-01&ed=2050-01-02&id=1", nil)
	ctx := getContext(request)
	request = request.WithContext(ctx)

	responseRecorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusSeeOther)
	}

	// second case -- database failed
	request, _ = http.NewRequest("GET", "/book-room?sd=2050-01-01&ed=2050-01-02&id=4", nil)
	ctx = getContext(request)
	request = request.WithContext(ctx)

	responseRecorder = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong response code trying to connect to access database: got %d, wanted %d", responseRecorder.Code, http.StatusSeeOther)
	}
}

func getContext(request *http.Request) context.Context {
	ctx, err := session.Load(request.Context(), request.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
