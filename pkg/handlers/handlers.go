package handlers

import (
	"net/http"

	"github.com/jackedelic/go-overview-trevor-sawler/pkg/render"
)

// Home handles home page
func Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "home.page.tmpl")
}

// About handles about page
func About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "about.page.tmpl")
}
