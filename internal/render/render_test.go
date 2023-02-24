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

func TestRenderTemplate(test *testing.T) {
	pathToTemplates = "./../../templates"
	tc, err := CreateTemplateCache()
	if err != nil {
		test.Error(err)
	}

	app.TemplateCache = tc

	sessionRequest, err := getSession()
	if err != nil {
		test.Error(err)
	}

	var writer myWriter

	err = Template(&writer, sessionRequest, "home.page.html", &models.TemplateData{})
	if err != nil {
		test.Error("error writing template to browser", err)
	}

	err = Template(&writer, sessionRequest, "non-existent.page.html", &models.TemplateData{})
	if err == nil {
		test.Error("rendered template that does not exist")
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

func TestNewTemplates(t *testing.T) {
	NewRenderer(app)
}

func TestCreateTemplateCache(test *testing.T) {
	pathToTemplates = "./../../templates"

	_, err := CreateTemplateCache()
	if err != nil {
		test.Error(err)
	}
}
