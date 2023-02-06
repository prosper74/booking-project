package render

import (
	"net/http"
	"testing"

	"github.com/atuprosper/booking-project/internal/models"
)

func TestAddDefaultData(test *testing.T) {
	var templateData models.TemplateData

	sessionRequest, err := getSession()
	if err != nil {
		test.Fatal(err)
	}

	session.Put(sessionRequest.Context(), "flash", "123")

	result := AddDefaultData(&templateData, sessionRequest)
	if result.Flash != "123" {
		test.Error("flash value of 123 not found in session")
	}

}

func getSession() (*http.Request, error) {
	request, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	requestContext := request.Context()
	requestContext, _ = session.Load(requestContext, request.Header.Get("X-Session"))
	request = request.WithContext(requestContext)
	return request, nil
}
