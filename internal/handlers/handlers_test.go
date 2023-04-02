package handlers

import (
	"context"
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
	{"alpine", "/alpine", "GET", http.StatusOK},
	{"generals", "/generals", "GET", http.StatusOK},
	{"reservation", "/reservation", "GET", http.StatusOK},
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

func TestRepository_MakeReservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Generals Suit",
		},
	}

	request, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getContext(request)
	request = request.WithContext(ctx)
	responseRecorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.MakeReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", responseRecorder.Code, http.StatusOK)
	}

	// test case where reeservation is not in session (reset everything)
	request, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getContext(request)
	request = request.WithContext(ctx)
	responseRecorder = httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code for reservation not in session: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// test with non-existent room
	request, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getContext(request)
	request = request.WithContext(ctx)
	responseRecorder = httptest.NewRecorder()

	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code for non existing room: got %d, wanted %d", responseRecorder.Code, http.StatusOK)
	}
}

func TestRepository_PostMakeReservation(t *testing.T) {
	postedData := url.Values{}
	postedData.Add("start_date", "2050-01-01")
	postedData.Add("end_date", "2050-01-02")
	postedData.Add("first_name", "Prosper")
	postedData.Add("last_name", "Atu")
	postedData.Add("email", "prosper@atu.com")
	postedData.Add("phone", "2255887744")
	postedData.Add("room_id", "1")

	request, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx := getContext(request)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder := httptest.NewRecorder()

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

func getContext(request *http.Request) context.Context {
	ctx, err := session.Load(request.Context(), request.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
