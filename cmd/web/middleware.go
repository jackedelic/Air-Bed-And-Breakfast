package main

import (
	"fmt"
	"net/http"

	"github.com/jackedelic/bookings/helpers"
	"github.com/justinas/nosurf"
)

// WriteToConsole is an http middleware that write to the fmt before calling the next middleware
func WriteToConsole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hit the page")
		next.ServeHTTP(w, r)
	})
}

// NoSurf is an http middleware that attaches cookie to "/"
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads the session from app.AppConfig
func SessionLoad(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next) // It remembers session based on X-Session header token, independent of IP address
}

// Auth authenticates the request's session (by checking for user_id in the session data for this session).
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "Log in first")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
