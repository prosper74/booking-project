package handlers

import (
	"fmt"
	"net/http"

	"github.com/atuprosper/booking-project/pkg/config"
	"github.com/atuprosper/booking-project/pkg/models"
	"github.com/atuprosper/booking-project/pkg/render"
)

// Creating a Repository pattern
// This variable is the repository used by the handlers
var Repo *Repository

type Repository struct {
	App *config.AppConfig
}

// This function creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// This function NewHandlers, sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// A function with a 'reciever' m, of type 'Repository'. This will give our handler function access to everything in the config file
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	//Perform some logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello again."

	getRemoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = getRemoteIP
	fmt.Println("Your IP address is", getRemoteIP)

	// Send the data to the template
	render.RenderTemplate(w, "home.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "about.page.html", &models.TemplateData{})
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "contact.page.html", &models.TemplateData{})
}

func (m *Repository) Alpine(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "alpine.page.html", &models.TemplateData{})
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "generals.page.html", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "reservation.page.html", &models.TemplateData{})
}
