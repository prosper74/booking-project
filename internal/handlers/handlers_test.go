package handlers

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

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
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code for non existing room: got %d, wanted %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func getContext(request *http.Request) context.Context {
	ctx, err := session.Load(request.Context(), request.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
