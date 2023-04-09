package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/atuprosper/booking-project/internal/config"
	"github.com/atuprosper/booking-project/internal/driver"
	"github.com/atuprosper/booking-project/internal/forms"
	"github.com/atuprosper/booking-project/internal/helpers"
	"github.com/atuprosper/booking-project/internal/models"
	"github.com/atuprosper/booking-project/internal/render"
	"github.com/atuprosper/booking-project/internal/repository"
	"github.com/atuprosper/booking-project/internal/repository/dbrepo"
)

var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// This function creates a new repository
func NewRepo(appConfig *config.AppConfig, dbConnectionPool *driver.DB) *Repository {
	return &Repository{
		App: appConfig,
		DB:  dbrepo.NewPostgresRepo(dbConnectionPool.SQL, appConfig),
	}
}

// This function creates a new repository
func NewTestRepo(appConfig *config.AppConfig) *Repository {
	return &Repository{
		App: appConfig,
		DB:  dbrepo.NewTestRepo(appConfig),
	}
}

// This function NewHandlers, sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// This function handles the Home page and renders the template
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.html", &models.TemplateData{})
}

// This function handles the About page and renders the template
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.html", &models.TemplateData{})
}

// This function handles the Contact page and renders the template
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.html", &models.TemplateData{})
}

// This function handles the single room(Luxery) page and renders the template
func (m *Repository) Alpine(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "alpine.page.html", &models.TemplateData{})
}

// This function handles the single room(Generals) page and renders the template
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.html", &models.TemplateData{})
}

// This function handles the reservation page and renders the template
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "reservation.page.html", &models.TemplateData{})
}

// This function handles the search page and displays the available rooms template
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	startDate, err := time.Parse("02-01-2006", r.Form.Get("start"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse("02-01-2006", r.Form.Get("end"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	for _, item := range rooms {
		m.App.InfoLog.Println("ROOM:", item.ID, item.RoomName)
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availabe rooms on the date selected")
		http.Redirect(w, r, "/reservation", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	reservationDates := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", reservationDates)

	render.Template(w, r, "available-rooms.page.html", &models.TemplateData{
		Data: data,
	})
}

// Availability json, to handle availability request and send back json
type jsonResponse struct {
	Ok        bool   `json:"ok"`
	Message   string `json:"message"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	RoomID    string `json:"room_id"`
}

// This function checks if the dates entered in a single room search has availability
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		response := jsonResponse{
			Ok:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(response, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	startDate, err := time.Parse("02-01-2006", r.Form.Get("start"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse("02-01-2006", r.Form.Get("end"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	roomId, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomId)
	if err != nil {
		response := jsonResponse{
			Ok:      false,
			Message: "Error connecting to the database",
		}

		out, _ := json.MarshalIndent(response, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	response := jsonResponse{
		Ok:        available,
		Message:   "",
		StartDate: r.Form.Get("start"),
		EndDate:   r.Form.Get("end"),
		RoomID:    r.Form.Get("room_id"),
	}

	out, _ := json.MarshalIndent(response, "", "    ")

	// Tell the browser the type of file in the header
	w.Header().Set("Content-Type", "Application/json")
	w.Write(out)
}

// This function handles the make reservation page and renders the template
func (m *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	reservationInSession, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(reservationInSession.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservationInSession.Room.RoomName = room.RoomName

	startDate := reservationInSession.StartDate.Format("2006-01-02")
	endDate := reservationInSession.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = startDate
	stringMap["end_date"] = endDate

	data := make(map[string]interface{})
	data["reservation"] = reservationInSession

	m.App.Session.Put(r.Context(), "reservation", reservationInSession)

	render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// This function POST the reservation and store them in the database
func (m *Repository) PostMakeReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	form := forms.New(r.PostForm)

	// Form validations
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3, 30)
	form.MinLength("last_name", 3, 30)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		m.App.Session.Put(r.Context(), "error", "Invalid form input")
		render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationId, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't insert into database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationId,
		RestrictionID: 1,
	}

	log.Println(reservation.StartDate)

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't insert room restriction!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Send email notification to customer
	htmlBody := fmt.Sprintf(`
	<strong>Thank you for making a reservation</strong><br />
	<p>Dear %s, </p>
	<p>This is to confirm your reservation from %s, to %s. </p>
	<p>We hope to see you soon</p>
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	message := models.MailData{
		To:       reservation.Email,
		From:     "me@prosper.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlBody,
		Template: "basic.html",
	}

	m.App.MailChannel <- message

	// Send email notification to admin
	htmlBody = fmt.Sprintf(`
	<strong>Hello, Admin</strong><br />
	<p>There is a new reservation from %s %s, </p>
	<p>Reservation Dates: %s, to %s. </p>
	<p>Room: %s. </p>
	<p>Customer Email: %s</p>
	`, reservation.FirstName, reservation.LastName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"), reservation.Room.RoomName, reservation.Email)

	message = models.MailData{
		To:      "Admin@email.com",
		From:    "Admin@server.com",
		Subject: "New Reservation",
		Content: htmlBody,
	}

	m.App.MailChannel <- message

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// This function handles the selected room from the available rooms displayed in search availability
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// split the URL up by /, and grab the 3rd element
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get the reservation in session
	reservationInSession, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Update the reservation by adding the room ID
	reservationInSession.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", reservationInSession)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// This function takes URL parameters, builds a sessional variable, and redirects user to make reservation page
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	var reservation models.Reservation
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	startDate, err := time.Parse("02-01-2006", r.URL.Query().Get("sd"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse("02-01-2006", r.URL.Query().Get("ed"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.Room.RoomName = room.RoomName
	reservation.RoomID = roomID
	reservation.StartDate = startDate
	reservation.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// This function displays the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok || reservation.FirstName == "" || reservation.LastName == "" || reservation.Email == "" || reservation.Phone == "" {
		m.App.ErrorLog.Println("Cannot get reservation from session")
		m.App.Session.Put(r.Context(), "error", "<h5>Can't get reservation from session!</h5><br /> Please, select an available room and make a reservation")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	startDate := reservation.StartDate.Format("2006-01-02")
	endDate := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = startDate
	stringMap["end_date"] = endDate

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})

	m.App.Session.Remove(r.Context(), "reservation")
}
