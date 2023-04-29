package main

import (
	"net/http"

	"github.com/atuprosper/booking-project/internal/config"
	"github.com/atuprosper/booking-project/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	// Add all our middlewares here
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Get("/rooms/{id}", handlers.Repo.SingleRoom)

	mux.Get("/reservation", handlers.Repo.Reservation)
	mux.Post("/reservation", handlers.Repo.PostReservation)
	mux.Post("/reservation-json", handlers.Repo.AvailabilityJSON)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/make-reservation", handlers.Repo.MakeReservation)
	mux.Post("/make-reservation", handlers.Repo.PostMakeReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	mux.Get("/user/login", handlers.Repo.Login)
	mux.Post("/user/login", handlers.Repo.PostLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		// Use the Auth middleware
		mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.AdminDashboard)

		mux.Get("/new-reservations", handlers.Repo.AdminNewReservations)
		mux.Get("/all-reservations", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handlers.Repo.AdminPostReservationsCalendar)

		mux.Get("/reservations/{src}/{id}/show", handlers.Repo.AdminSingleReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.PostAdminSingleReservation)

		mux.Get("/rooms", handlers.Repo.AdminAllRooms)
		mux.Get("/rooms/{id}", handlers.Repo.AdminSingleRoom)
		mux.Post("/rooms/{id}", handlers.Repo.PostAdminSingleRoom)
		mux.Get("/rooms/new-room", handlers.Repo.AdminNewRoom)
		mux.Post("/rooms/new-room", handlers.Repo.PostAdminNewRoom)
		mux.Get("/delete-room/{id}", handlers.Repo.AdminDeleteRoom)

		mux.Get("/process-reservation/{src}/{id}/do", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}/do", handlers.Repo.AdminDeleteReservation)

		mux.Get("/todo-list", handlers.Repo.AdminTodoList)
		mux.Post("/todo-list", handlers.Repo.PostAdminTodoList)
		mux.Get("/delete-todo/{id}", handlers.Repo.AdminDeleteTodo)
	})

	return mux
}
