package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/atuprosper/booking-project/internal/config"
	"github.com/atuprosper/booking-project/internal/forms"
	"github.com/atuprosper/booking-project/internal/models"
	"github.com/atuprosper/booking-project/internal/render"
)

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
	render.RenderTemplate(w, r, "home.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "about.page.html", &models.TemplateData{})
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.html", &models.TemplateData{})
}

func (m *Repository) Alpine(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "alpine.page.html", &models.TemplateData{})
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.html", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "reservation.page.html", &models.TemplateData{})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	w.Write([]byte(fmt.Sprintf("Start date is %s, End date is %s", start, end)))
}

// Availability json, to handle availability request and send back json
type jsonResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		Ok:      true,
		Message: "Available!",
	}

	out, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		log.Println(err)
	}

	// Tell the browser the type of file in the header
	w.Header().Set("Content-Type", "Application/json")
	w.Write(out)
}

func (m *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	var reservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "make-reservation.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) PostMakeReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)

	// Form validations
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3, 30)
	form.MinLength("last_name", 3, 30)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.RenderTemplate(w, r, "make-reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		log.Println("Cannot get reservation from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data: data,
	})
}
