package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// We can write our own middleware for 'chi' package in this file.

// 'next' is commonly used as the parameter for custome middleware

// This middleware adds CSRF protection to all POST request
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InPrduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// This middleware makes our server to be 'state' aware, in order to keep our session in state
func SessionLoad(next http.Handler) http.Handler {
	// LoadAndSave provides middleware which automatically loads and saves session data for the current request, and communicates the session token to and from the client in a cookie.
	return session.LoadAndSave(next)
}
