package handlers

import (
	"net/http"

	"github.com/jackedelic/go-overview-trevor-sawler/pkg/config"
	"github.com/jackedelic/go-overview-trevor-sawler/pkg/render"
)

// TemplateData holds data sent from handler to template
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float32
	Data      map[string]interface{}
	CSRFToken string
	Flash     string
	Warning   string
	Error     string
}

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
	render.RenderTemplate(w, "home.page.tmpl")
}

// About handles about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "about.page.tmpl")
}
