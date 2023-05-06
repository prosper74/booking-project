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

// TestPostReservation tests the PostReservation handler
func TestPostReservation(t *testing.T) {
	for _, e := range postReservationTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)

		}
		ctx := getContext(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostReservation)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedResponseCode)
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

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

// testAvailabilityJSONData is data for the AvailabilityJSON handler, /reservation-json route
var testAvailabilityJSONData = []struct {
	name            string
	postedData      url.Values
	expectedOK      bool
	expectedMessage string
}{
	{
		name: "rooms not available",
		postedData: url.Values{
			"start":   {"2050-01-01"},
			"end":     {"2050-01-02"},
			"room_id": {"1"},
		},
		expectedOK: false,
	}, {
		name: "rooms are available",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedOK: true,
	},
	{
		name:            "empty post body",
		postedData:      nil,
		expectedOK:      false,
		expectedMessage: "Internal Server Error",
	},
	{
		name: "database query fails",
		postedData: url.Values{
			"start":   {"2060-01-01"},
			"end":     {"2060-01-02"},
			"room_id": {"1"},
		},
		expectedOK:      false,
		expectedMessage: "Error querying database",
	},
}

// TestAvailabilityJSON tests the AvailabilityJSON handler
func TestAvailabilityJSON(t *testing.T) {
	for _, e := range testAvailabilityJSONData {
		// create request, get the context with session, set header, create recorder
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/reservation-json", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/reservation-json", nil)
		}
		ctx := getContext(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// make our handler a http.HandlerFunc and call
		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(rr, req)

		var j jsonResponse
		err := json.Unmarshal([]byte(rr.Body.String()), &j)
		if err != nil {
			t.Error("failed to parse json!")
		}

		if j.Ok != e.expectedOK {
			t.Errorf("%s: expected %v but got %v", e.name, e.expectedOK, j.Ok)
		}
	}
}

// testPostAvailabilityData is data for the PostAvailability handler test, /make-reservation
var testPostAvailabilityData = []struct {
	name               string
	postedData         url.Values
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "rooms not available",
		postedData: url.Values{
			"start": {"2050-01-01"},
			"end":   {"2050-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "rooms are available",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "empty post body",
		postedData:         url.Values{},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "start date wrong format",
		postedData: url.Values{
			"start":   {"invalid"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "end date wrong format",
		postedData: url.Values{
			"start": {"2040-01-01"},
			"end":   {"invalid"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "database query fails",
		postedData: url.Values{
			"start": {"2060-01-01"},
			"end":   {"2060-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
}

// TestPostAvailability tests the PostAvailabilityHandler
func TestPostAvailability(t *testing.T) {
	for _, e := range testPostAvailabilityData {
		req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(e.postedData.Encode()))

		// get the context with session
		ctx := getContext(req)
		req = req.WithContext(ctx)

		// set the request header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// make our handler a http.HandlerFunc and call
		handler := http.HandlerFunc(Repo.PostMakeReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s gave wrong status code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

// reservationSummaryTests is the data to test ReservationSummary handler
var reservationSummaryTests = []struct {
	name               string
	reservation        models.Reservation
	url                string
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "res-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "Generals Suit",
			},
		},
		url:                "/reservation-summary",
		expectedStatusCode: http.StatusOK,
		expectedLocation:   "",
	},
	{
		name:               "res-not-in-session",
		reservation:        models.Reservation{},
		url:                "/reservation-summary",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

// TestReservationSummary tests the ReservationSummaryHandler
func TestReservationSummary(t *testing.T) {
	for _, e := range reservationSummaryTests {
		req, _ := http.NewRequest("GET", e.url, nil)
		ctx := getContext(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.ReservationSummary)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}
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
