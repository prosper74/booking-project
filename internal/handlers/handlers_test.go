package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
	// {"make-reservation", "/make-reservation", "GET", []postData{}, http.StatusOK},
	// {"reservation-summary", "/reservation-summary", "GET", []postData{}, http.StatusOK},
	// {"post-reservation", "/reservation", "POST", []postData{
	// 	{key: "start", value: "2023-01-01"},
	// 	{key: "end", value: "2023-01-02"},
	// }, http.StatusOK},
	// {"post-reservation-json", "/reservation-json", "POST", []postData{
	// 	{key: "start", value: "2023-01-01"},
	// 	{key: "end", value: "2023-01-02"},
	// }, http.StatusOK},
	// {"make-reservation", "/make-reservation", "POST", []postData{
	// 	{key: "first_name", value: "Prosper"},
	// 	{key: "last_name", value: "Atu"},
	// 	{key: "email", value: "me@email.com"},
	// 	{key: "phone", value: "555-555-5555"},
	// }, http.StatusOK},
}

func TestHandlers(testPointer *testing.T) {
	routes := getRoutes()

	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close()

	for _, element := range theTests {
		resp, err := testServer.Client().Get(testServer.URL + element.url)
		if err != nil {
			testPointer.Log(err)
			testPointer.Fatal(err)
		}

		if resp.StatusCode != element.expectedStatusCode {
			testPointer.Errorf("for %s expected %d but got %d", element.name, element.expectedStatusCode, resp.StatusCode)
		}

	}
}
