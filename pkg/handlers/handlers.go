package handlers

import (
	"fmt"
	"net/http"

	"github.com/jackedelic/bookings/pkg/config"
	"github.com/jackedelic/bookings/pkg/models"
	"github.com/jackedelic/bookings/pkg/render"
)

// Repo is a pointer to a Repository
var Repo *Repository

// Repository is a struct that holds AppConfig
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a pointer to a Repository using AppConfig passed to the function.
func NewRepo(appConfig *config.AppConfig) *Repository {
	repo := Repository{
		App: appConfig,
	}
	return &repo
}

// NewHandlers assigns the input repo to handlers.Repo
func NewHandlers(repo *Repository) {
	Repo = repo
}

// Home handles home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	m.App.Session.Put(r.Context(), "remote_ip", r.RemoteAddr)
	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
}

// About handles about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{"test": "Hello Again"}
	fmt.Println(r.Context())
	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}
