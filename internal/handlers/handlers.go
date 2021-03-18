package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jackedelic/bookings/forms"
	"github.com/jackedelic/bookings/internal/config"
	"github.com/jackedelic/bookings/internal/models"
	"github.com/jackedelic/bookings/internal/render"
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
	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About handles about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{"test": "Hello Again"}
	fmt.Println(r.Context())
	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// MakeReservation handles GET /make-reservation
func (m *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	emptyReservation := models.Reservation{}
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation
	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// PostReservation handles POST /post-reservation
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)
	log.Println(r.PostForm)
	log.Println(r.Form)
	// form.Has("first_name", r)

	form.Required("email", "first_name", "last_name")
	form.MinLength("first_name", 3, r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}
}

// GeneralsQuarters handles /generals-quarter
func (m *Repository) GeneralsQuarters(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	render.RenderTemplate(w, r, "generals-quarters.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// MajorsSuite handles /major-suite
func (m *Repository) MajorsSuite(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	render.RenderTemplate(w, r, "majors-suite.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// SearchAvailability handles /search-availability
func (m *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// PostSearchAvailability handles POST /search-availability
func (m *Repository) PostSearchAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	w.Write([]byte(fmt.Sprintf("star tdate is %s, end date is %s", start, end)))
}

// Contact handles /contact
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// ReceiveJSON handles POST /receive-json
func (m *Repository) ReceiveJSON(w http.ResponseWriter, r *http.Request) {
	type JSONResponse struct {
		Ok      bool   `json:"ok"`
		Message string `json:"message"`
	}
	jr := JSONResponse{Ok: true, Message: "You are in /receive-json"}
	jByte, err := json.MarshalIndent(jr, "", "    ")
	if err != nil {
		w.Write([]byte("Got error marshalling the json response"))
	}
	log.Println(string(jByte))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jByte)
}
