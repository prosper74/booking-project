package main

import (
	"fmt"
	"testing"

	"github.com/atuprosper/booking-project/internal/config"
	"github.com/go-chi/chi/v5"
)

func TestRoutes(testPointer *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch handlerType := mux.(type) {
	case *chi.Mux:
		// do nothing
	default:
		testPointer.Error(fmt.Sprintf("type is not *chi.Mux, but is %T", handlerType))
	}
}
