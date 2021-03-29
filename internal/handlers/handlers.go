package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackedelic/bookings/driver"
	"github.com/jackedelic/bookings/forms"
	"github.com/jackedelic/bookings/helpers"
	"github.com/jackedelic/bookings/internal/config"
	"github.com/jackedelic/bookings/internal/models"
	"github.com/jackedelic/bookings/internal/render"
	"github.com/jackedelic/bookings/repository"
	"github.com/jackedelic/bookings/repository/dbrepo"
)

// Repo is a pointer to a Repository
var Repo *Repository

// Repository is a struct that holds AppConfig
type Repository struct {
	App    *config.AppConfig
	DBRepo repository.DatabaseRepo
}

// NewRepo creates a pointer to a Repository using AppConfig passed to the function.
func NewRepo(appConfig *config.AppConfig, db driver.DB) *Repository {
	repo := Repository{
		App:    appConfig,
		DBRepo: dbrepo.NewPostgresRepo(db.SQL, appConfig),
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
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About handles about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{"test": "Hello Again"}
	fmt.Println(r.Context())
	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	render.Template(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// MakeReservation handles GET /make-reservation
func (m *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	emptyReservation := models.Reservation{}
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation
	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// PostReservation handles POST /make-reservation
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	sdStr := r.Form.Get("start_date")
	sd, err := time.Parse("2006-01-02", sdStr)
	if err != nil {
		helpers.ServerError(w, err)
	}
	edStr := r.Form.Get("end_date")
	ed, err := time.Parse("2006-01-02", edStr)
	if err != nil {
		helpers.ServerError(w, err)
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Creates Reservation object (corresponds to a row in a db)
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		StartDate: sd,
		EndDate:   ed,
		RoomID:    roomID,
	}

	form := forms.New(r.PostForm)
	log.Println(r.PostForm)
	log.Println(r.Form)

	form.Required("email", "first_name", "last_name")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DBRepo.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Creates a RoomRestriction object (corresponds to a row in a db)
	roomRestriction := models.RoomRestriction{
		StartDate:     sd,
		EndDate:       ed,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1, // Reservation type
	}

	err = m.DBRepo.InsertRoomRestriction(roomRestriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Stores the form data into our session storage (in-memory by default)
	m.App.Session.Put(r.Context(), "reservation", reservation) // The Session manager takes the session data from
	// this request's Context (session data is loaded by Session middleware earlier in the handler chain),
	// and stores our reservation -> sd.values["reservation"] = reservation
	// session data contains the token taken from the request header cookie, or generated by Session middleware.
	// sd := &sessionData{
	// 	status: Unmodified,
	// 	token:  token,
	// }
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// ReservationSummary handles GET /reservation-summary
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	// Retrieves the form data from session
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation) // reason for -> gob.Register(models.Reservation{}) in main.go
	if !ok {
		log.Println("Cannot get item from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = reservation
	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
	// Remove the reservation data from the session
	m.App.Session.Remove(r.Context(), "reservation")
}

// GeneralsQuarters handles /generals-quarters
func (m *Repository) GeneralsQuarters(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	render.Template(w, r, "generals-quarters.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// MajorsSuite handles /major-suite
func (m *Repository) MajorsSuite(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	render.Template(w, r, "majors-suite.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// SearchAvailability handles /search-availability
func (m *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// PostSearchAvailability handles POST /search-availability
func (m *Repository) PostSearchAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	sd, err := time.Parse("02-01-2006", start)
	if err != nil {
		helpers.ServerError(w, err)
	}
	ed, err := time.Parse("02-01-2006", end)
	if err != nil {
		helpers.ServerError(w, err)
	}
	rooms, err := m.DBRepo.SearchAvailableRoomsByDates(sd, ed)
	if err != nil {
		helpers.ServerError(w, err)
	}

	for _, r := range rooms {
		m.App.InfoLog.Println("ROOM:", r.ID, r.RoomName)
	}
	// No room available for the given date range
	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability")
		// w.Write([]byte(fmt.Sprintf("star tdate is %s, end date is %s", start, end)))
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	// For use in template data
	data := make(map[string]interface{})
	data["rooms"] = rooms
	// Store startdate and enddate in session for use in redirection
	res := models.Reservation{
		StartDate: sd,
		EndDate:   ed,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	// w.Write([]byte(fmt.Sprintf("star tdate is %s, end date is %s", start, end)))
	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// Contact handles /contact
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{}
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{
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
		helpers.ServerError(w, err)
		w.Write([]byte("Got error marshalling the json response"))
	}
	log.Println(string(jByte))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jByte)
}

// ChooseRoom retrieves reservation value from the session and converts it back to models.Reservation,
// and appends the choosen roomID (from url param when user click the link).
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation) // in main.go: gob.Register(models.Reservation{})
	if !ok {
		helpers.ServerError(w, err)
		return
	}
	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
