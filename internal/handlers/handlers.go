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
	"github.com/go-chi/chi/v5"
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
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms
	render.Template(w, r, "home.page.html", &models.TemplateData{
		Data: data,
	})
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
func (m *Repository) SingleRoom(w http.ResponseWriter, r *http.Request) {
	urlParams := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(urlParams[2])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room, err := m.DB.GetRoomByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["room"] = room

	render.Template(w, r, "single-room.page.html", &models.TemplateData{
		Data: data,
	})
}

// This function handles the reservation page and renders the template
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "reservation.page.html", &models.TemplateData{})
}

// This function handles the search page and displays the available rooms template
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	startDate, err := time.Parse("2006-01-02", r.Form.Get("start"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse("2006-01-02", r.Form.Get("end"))
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

	startDate, err := time.Parse("2006-01-02", r.Form.Get("start"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse("2006-01-02", r.Form.Get("end"))
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
		From:     "prosperdevstack@gmail.com",
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
		To:      "atu.prosper@gmail.com",
		From:    "prosperdevstack@gmail.com",
		Subject: "New Reservation",
		Content: htmlBody,
	}

	m.App.MailChannel <- message
	// End of emails

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
	startDate, err := time.Parse("2006-01-02", r.URL.Query().Get("sd"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse("2006-01-02", r.URL.Query().Get("ed"))
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

// This function handles the Admin Login page and renders the template
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	userExists := m.App.Session.GetInt(r.Context(), "user_id")
	if userExists > 0 {
		m.App.Session.Put(r.Context(), "warning", "You are already logged in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	render.Template(w, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// This function handles user details and authentication
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	// Allways renew the token in seesion for login or logout
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		stringMap := make(map[string]string)
		stringMap["email"] = email
		m.App.Session.Put(r.Context(), "error", "Invalid inputs")
		render.Template(w, r, "login.page.html", &models.TemplateData{
			Form:      form,
			StringMap: stringMap,
		})

		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)

		m.App.Session.Put(r.Context(), "error", "Invalid email/password")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)

		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Login Successful")
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// This function logs out the user
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	m.App.Session.Put(r.Context(), "warning", "You have logged out of your account")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handles the admin dashborad
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.html", &models.TemplateData{})
}

// Handles the new-reservations route
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-new-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

// Handles the all-reservations route
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

// Handles the single-reservation route
func (m *Repository) AdminSingleReservation(w http.ResponseWriter, r *http.Request) {
	urlParams := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(urlParams[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := urlParams[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")
	stringMap["year"] = year
	stringMap["month"] = month

	reservation, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "admin-single-reservation.page.html", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// Handles the single-reservation route for POST
func (m *Repository) PostAdminSingleReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	urlParams := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(urlParams[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := urlParams[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	reservation, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3, 30)
	form.MinLength("last_name", 3, 30)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		m.App.Session.Put(r.Context(), "error", "Invalid inputs")
		render.Template(w, r, "admin-single-reservation.page.html", &models.TemplateData{
			Form:      form,
			Data:      data,
			StringMap: stringMap,
		})

		return
	}

	err = m.DB.UpdateReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.Form.Get("year")
	month := r.Form.Get("month")

	m.App.Session.Put(r.Context(), "flash", "Reservation Updated")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/%s-reservations", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// Handles the processing of revervation
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	reservation, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if reservation.Processed == 0 {
		err = m.DB.UpdateProcessedForReservation(id, 1)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		m.App.Session.Put(r.Context(), "flash", "<strong>Successful!!!</strong><br><br> <p>Reservation is now marked as Processed</p>")
	} else {
		err = m.DB.UpdateProcessedForReservation(id, 0)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		m.App.Session.Put(r.Context(), "flash", "<strong>Successful!!!</strong><br><br> <p>Reservation is now marked as Not Processed</p>")
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/%s-reservations", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

// Handles the deleting of revervation
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.DeleteReservation(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "<strong>Successful!!!</strong><br><br> <p>Reservation Deleted</p>")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/%s-reservations", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// Handles the reservations-calendar route
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))

		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	// get next and previous months/year
	nextMonth := now.AddDate(0, 1, 0)
	previousMonth := now.AddDate(0, -1, 0)

	stringMap := make(map[string]string)
	stringMap["this_month"] = now.Format("01")
	stringMap["this_year"] = now.Format("2006")
	stringMap["next_month"] = nextMonth.Format("01")
	stringMap["next_year"] = nextMonth.Format("2006")
	stringMap["previous_month"] = previousMonth.Format("01")
	stringMap["previous_year"] = previousMonth.Format("2006")

	// Get the first and last day of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {
		// get the reservation and block maps
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		// get all the restrictions for the current room
		restrictions, err := m.DB.GetRestrictionsForCurrentRoom(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range restrictions {
			if y.ReservationID > 0 {
				// it's a reservation
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			} else {
				// it's a block
				blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			}
		}
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.html", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// Handles the reservation calendar POST route
func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("year"))
	month, _ := strconv.Atoi(r.Form.Get("month"))

	//Process changes
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	for _, x := range rooms {
		// Get the block map from the session. Loop through entire map, if we have an entry in the map
		// that does not exist in our posted data, and if the restriction id > 0, then it is a block we need to
		// remove.
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for name, value := range curMap {
			// ok will be false if the value is not in the map
			if val, ok := curMap[name]; ok {
				// only pay attention to values > 0, and that are not in the form post
				// the rest are just placeholders for days without blocks
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						// delete the restriction by id
						err := m.DB.DeleteBlockByID(value)
						if err != nil {
							log.Println(err)
						}
						m.App.Session.Put(r.Context(), "flash", "Block removed successfully")
					}
				}
			}
		}
	}

	// now handle new blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-02", exploded[3])
			// insert a new block
			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				log.Println(err)
			}
			m.App.Session.Put(r.Context(), "flash", "Reservation Block Updated")
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}

// Handles the all-rooms route
func (m *Repository) AdminAllRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	render.Template(w, r, "admin-all-rooms.page.html", &models.TemplateData{
		Data: data,
	})
}

// Handles the single-room route
func (m *Repository) AdminSingleRoom(w http.ResponseWriter, r *http.Request) {
	urlParams := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(urlParams[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room, err := m.DB.GetRoomByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["room"] = room

	render.Template(w, r, "admin-single-room.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// Handles the single-room route for POST
func (m *Repository) PostAdminSingleRoom(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	urlParams := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(urlParams[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room, err := m.DB.GetRoomByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room.RoomName = r.Form.Get("room_name")
	room.Price = r.Form.Get("price")
	room.ImageSource = r.Form.Get("image_src")
	room.Description = r.Form.Get("description")

	form := forms.New(r.PostForm)
	form.Required("room_name", "price", "image_src", "description")
	form.MinLength("room_name", 5, 30)
	form.MinLength("description", 5, 20000)

	if !form.Valid() {
		data := make(map[string]interface{})
		data["room"] = room
		m.App.Session.Put(r.Context(), "error", "Invalid inputs")
		render.Template(w, r, "admin-single-room.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	err = m.DB.UpdateRoom(room)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Reservation Updated")
	http.Redirect(w, r, "/admin/rooms", http.StatusSeeOther)
}

// Handles the new-room route to create a new room
func (m *Repository) AdminNewRoom(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-new-room.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// This function POST the new room form and store them in the database
func (m *Repository) PostAdminNewRoom(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var room models.Room

	room.RoomName = r.Form.Get("room_name")
	room.Price = r.Form.Get("price")
	room.ImageSource = r.Form.Get("image_src")
	room.Description = r.Form.Get("description")

	// Form validations
	form := forms.New(r.PostForm)
	form.Required("room_name", "price", "image_src", "description")
	form.MinLength("room_name", 5, 30)
	form.MinLength("description", 5, 20000)

	if !form.Valid() {
		data := make(map[string]interface{})
		data["room"] = room
		m.App.Session.Put(r.Context(), "error", "Invalid form input")
		render.Template(w, r, "admin-new-room.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Insert new room here
	err = m.DB.InsertRoom(room)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't insert into database")
		helpers.ServerError(w, err)
		http.Redirect(w, r, "/admin/rooms", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Room Created Successfully!!!")
	http.Redirect(w, r, "/admin/rooms", http.StatusSeeOther)
}

func (m *Repository) AdminDeleteRoom(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	err := m.DB.DeleteRoom(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "<strong>Successful!!!</strong><br><br> <p>Room Deleted</p>")
	http.Redirect(w, r, "/admin/rooms", http.StatusSeeOther)
}

// Handles the admin todo list route
func (m *Repository) AdminTodoList(w http.ResponseWriter, r *http.Request) {
	userID := m.App.Session.GetInt(r.Context(), "user_id")

	todoList, err := m.DB.GetTodoListByUserID(userID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["todo_list"] = todoList

	render.Template(w, r, "admin-todo.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// Handles the Post for todo list route
func (m *Repository) PostAdminTodoList(w http.ResponseWriter, r *http.Request) {
	userID := m.App.Session.GetInt(r.Context(), "user_id")

	var todoList models.TodoList

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	todoList.Todo = r.Form.Get("todo_list")
	todoList.UserID = userID

	form := forms.New(r.PostForm)
	form.Required("todo_list")
	form.MinLength("todo_list", 5, 255)

	if !form.Valid() {
		data := make(map[string]interface{})
		data["todo_list"] = todoList
		m.App.Session.Put(r.Context(), "error", "Invalid form input. <br> must be less than 250 and more than 5 characters")
		render.Template(w, r, "admin-todo.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Insert new todo here
	err = m.DB.InsertTodoList(todoList)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't insert into database")
		helpers.ServerError(w, err)
		http.Redirect(w, r, "/admin/todo-list", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Todo Created Successfully!!!")
	http.Redirect(w, r, "/admin/todo-list", http.StatusSeeOther)
}

// AdminDeleteTodo deletes Todo from the database and
func (m *Repository) AdminDeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	err := m.DB.DeleteTodo(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "<strong>Successful!!!</strong><br><br> <p>Todo Deleted</p>")
	http.Redirect(w, r, "/admin/todo-list", http.StatusSeeOther)
}
