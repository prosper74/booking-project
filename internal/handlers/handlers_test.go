package handlers

import (
	"net/http/httptest"
	"net/url"
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
	params             []postData
	expectedStatusCode int
}{
	// {"home", "/", "GET", []postData{}, http.StatusOK},
	// {"about", "/about", "GET", []postData{}, http.StatusOK},
	// {"contact", "/contact", "GET", []postData{}, http.StatusOK},
	// {"alpine", "/alpine", "GET", []postData{}, http.StatusOK},
	// {"generals", "/generals", "GET", []postData{}, http.StatusOK},
	// {"reservation", "/reservation", "GET", []postData{}, http.StatusOK},
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
		if element.method == "GET" {
			resp, err := testServer.Client().Get(testServer.URL + element.url)
			if err != nil {
				testPointer.Log(err)
				testPointer.Fatal(err)
			}

			if resp.StatusCode != element.expectedStatusCode {
				testPointer.Errorf("for %s expected %d but got %d", element.name, element.expectedStatusCode, resp.StatusCode)
			}
		} else {
			values := url.Values{}
			for _, item := range element.params {
				values.Add(item.key, item.value)
			}

			resp, err := testServer.Client().PostForm(testServer.URL+element.url, values)
			if err != nil {
				testPointer.Log(err)
				testPointer.Fatal(err)
			}

			if resp.StatusCode != element.expectedStatusCode {
				testPointer.Errorf("for %s expected %d but got %d", element.name, element.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}
