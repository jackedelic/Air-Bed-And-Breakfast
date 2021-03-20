package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackedelic/bookings/internal/config"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig
	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
	case http.Handler:
	default:
		t.Error(fmt.Sprintf("type is not myHttpHandler, but is %T", v))
	}
}
