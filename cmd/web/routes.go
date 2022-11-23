package main

import (
	"net/http"

	"github.com/atuprosper/booking-project/pkg/config"
	"github.com/atuprosper/booking-project/pkg/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	// Create a multiplexer 'mux'
	mux := chi.NewRouter()

	// Add all our middlewares here
	// Adding a 'Recoverer' middleware from chi package. It helps for panic
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)

	return mux
}
