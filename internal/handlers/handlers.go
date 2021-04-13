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

type JSONResponse struct {
	Ok        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// NewRepo creates a pointer to a Repository using AppConfig passed to the function.
func NewRepo(appConfig *config.AppConfig, db driver.DB) *Repository {
	repo := Repository{
		App:    appConfig,
		DBRepo: dbrepo.NewPostgresRepo(db.SQL, appConfig),
	}
	return &repo
}

// NewTestingRepo creates a pointer to a testingDBRepo using AppConfig.
func NewTestingRepo(appConfig *config.AppConfig) *Repository {
	repo := Repository{
		App:    appConfig,
		DBRepo: dbrepo.NewTestingRepo(appConfig),
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
	// Pull out reservation from session
	reserv, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Error retrieving reservation from session")
		m.App.Session.Put(r.Context(), "error", "error getting reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// convert the format of date into the format accepted by html date
	startDate := reserv.StartDate.Format("2006-01-02")
	endDate := reserv.EndDate.Format("2006-01-02")

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Data:      map[string]interface{}{"reservation": reserv},
		Form:      forms.New(nil),
		StringMap: map[string]string{"start_date": startDate, "end_date": endDate},
	})
}

// PostReservation handles POST /make-reservation
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("Error parsing form")
		m.App.Session.Put(r.Context(), "error", "error parsing form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.ErrorLog.Println("Error converting room_id to integer")
		m.App.Session.Put(r.Context(), "error", "error processing room_id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Pull reservation (partially filled with start,end dates) from session
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation) // RoomID could be populated already
	// Check type
	resType := fmt.Sprintf("%T", reservation)
	fmt.Println(resType)
	if !ok {
		m.App.ErrorLog.Println("Error retrieving reservation from session")
		m.App.Session.Put(r.Context(), "error", "error getting reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	if roomID != 0 { // room_id supplied via form. otherwise its from session
		reservation.RoomID = roomID
	}
	// reservation.Room and reservation.RoomID should be in the session already

	form := forms.New(r.PostForm)
	// log.Println(r.PostForm)
	log.Println(r.Form)

	form.Required("email", "first_name", "last_name")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		w.WriteHeader(http.StatusTemporaryRedirect)

		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}
	newReservationID, err := m.DBRepo.InsertReservation(reservation)
	if err != nil {
		m.App.ErrorLog.Println("Error inserting reservation to the database", err)
		m.App.Session.Put(r.Context(), "error", "server error inserting reservation the database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Creates a RoomRestriction object (corresponds to a row in a db)
	roomRestriction := models.RoomRestriction{
		StartDate:     reservation.StartDate, // dd-mm-yyyy (but mysql stores yyyy-mm-dd)
		EndDate:       reservation.EndDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1, // Reservation type
	}

	err = m.DBRepo.InsertRoomRestriction(roomRestriction)
	if err != nil {
		m.App.ErrorLog.Println("Error inserting room_restriction to the database")
		m.App.Session.Put(r.Context(), "error", "server error inserting room restriction to the database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Send email notification to customer
	htmlMsg := fmt.Sprintf(`
			<strong>Reservation Confirmation</strong><br>
			Dear %s, <br>
			This is to confirm your reservation from %s to %s
		`,
		reservation.FirstName,
		reservation.StartDate.Format("02-01-2006"),
		reservation.EndDate.Format("02-01-2006"))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "jackwong@airbnbdestroyer.com",
		Subject:  "Reservation confirmation",
		Content:  htmlMsg,
		Template: "basic.html",
	}
	m.App.MailChan <- msg
	// Send email to property owner
	htmlMsg = fmt.Sprintf(`
			<strong>Reservation Notification</strong>
			A reservation has been made for %s from %s to %s.
		`,
		reservation.Email,
		reservation.StartDate.Format("02-01-2006"),
		reservation.EndDate.Format("02-01-2006"))
	msg = models.MailData{
		To:       "jackwong@owner.com",
		From:     "your-program@bookings.com",
		Subject:  "Reservation Notification",
		Content:  htmlMsg,
		Template: "basic.html",
	}
	m.App.MailChan <- msg

	// Update session's reservation

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

	stringMap := make(map[string]string)
	stringMap["start_date"] = reservation.StartDate.Format("2006-01-02")
	stringMap["end_date"] = reservation.EndDate.Format("2006-01-02")

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
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

// SearchAvailabilityJSON handles POST /search-availability-json
func (m *Repository) SearchAvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("Error parsing form")
		resp := JSONResponse{
			Ok:      false,
			Message: "Internal server error parsing form",
		}
		out, _ := json.MarshalIndent(resp, "", "		")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}
	// Retrieves start date and end date from the form
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.ErrorLog.Println("Error converting room_id to integer")
		resp := JSONResponse{
			Ok:      false,
			Message: "Internal server error processing room_id",
		}
		out, _ := json.MarshalIndent(resp, "", "		")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}
	sd, err := time.Parse("02-01-2006", start)
	if err != nil {
		m.App.ErrorLog.Println("Error retrieving start_date from form")
		resp := JSONResponse{
			Ok:      false,
			Message: "Internal server error retrieving start_date from form",
		}
		out, _ := json.MarshalIndent(resp, "", "		")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}
	ed, err := time.Parse("02-01-2006", end)
	if err != nil {
		m.App.ErrorLog.Println("Error parsing end_date from form")
		resp := JSONResponse{
			Ok:      false,
			Message: "Internal server error parsing end_date from form",
		}
		out, _ := json.MarshalIndent(resp, "", "		")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}
	// Find if any available rooms from our database
	available, err := m.DBRepo.SearchAvailabilityByDatesByRoomID(sd, ed, roomID)
	if err != nil {
		m.App.ErrorLog.Println("Error searching availability by dates by room id")
		resp := JSONResponse{
			Ok:      false,
			Message: "Internal server error searching availability by dates by room id",
		}
		out, _ := json.MarshalIndent(resp, "", "		")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	m.App.InfoLog.Println(available)

	jr := JSONResponse{
		Ok:        available,
		Message:   "You are in /search-available-json",
		RoomID:    strconv.Itoa(roomID),
		StartDate: start,
		EndDate:   end,
	}
	jByte, _ := json.MarshalIndent(jr, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jByte)
}

// PostSearchAvailability handles POST /search-availability and look for available rooms/
// If rooms available, it puts them in template data and renders the template it to client.
func (m *Repository) PostSearchAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("Error parsing form")
		helpers.ServerError(w, err)
		return
	}
	// Retrieve start date and end date from the form
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

	// Get all available rooms from db for the given start and end dates.
	rooms, err := m.DBRepo.SearchAvailableRoomsByDates(sd, ed)
	if err != nil {
		helpers.ServerError(w, err)
	}

	for _, r := range rooms {
		m.App.InfoLog.Println("ROOM:", r.ID, r.RoomName)
	}
	// If no room available for the given date range
	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability")
		// w.Write([]byte(fmt.Sprintf("star tdate is %s, end date is %s", start, end)))
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}
	// If any room available, save it into a models.TemplateData object.
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

	// Get room from db and populates models.Reservation
	room, err := m.DBRepo.GetRoomById(roomID)
	if err != nil {
		helpers.ServerError(w, err)
	}
	res.Room = room

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes URL paramaters, builds a sessional variable and takes user to make res screen.
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// Retrieve room id, start date and end date from url query params
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	startDate := r.URL.Query().Get("s")
	endDate := r.URL.Query().Get("e")

	// Create models.Reservation object using the data retrieved
	var reservation models.Reservation

	sd, err := time.Parse("02-01-2006", startDate)
	if err != nil {
		helpers.ServerError(w, err)
	}
	reservation.StartDate = sd

	ed, err := time.Parse("02-01-2006", endDate)
	if err != nil {
		helpers.ServerError(w, err)
	}
	reservation.EndDate = ed

	room, err := m.DBRepo.GetRoomById(roomID)
	if err != nil {
		helpers.ServerError(w, err)
	}
	reservation.Room = room
	reservation.RoomID = roomID

	// Put out models.Reservatio object into session
	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// ShowLogin renders the login page
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostLogin authenticates the user-given email and password
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println(err)
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		w.WriteHeader(http.StatusUnauthorized)
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DBRepo.Authenticate(email, password)
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout destroys the session data and assigns new session token to the user
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// AdminDashboard renders the admin-dashboard page
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminNewReservations renders admin page with all new reservations data
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DBRepo.GetAllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminAllReservations renders admin page with all reservations data
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DBRepo.GetAllReservations()
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Internal error getting all reservations")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation renders reservation page for a particular reservation
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	resID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error parsing url param")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	src := chi.URLParam(r, "src")
	stringMap := make(map[string]string)
	stringMap["src"] = src

	// Get reservation from the database
	res, err := m.DBRepo.GetReservationByID(resID)
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Internal error fetching reservation")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	m.App.InfoLog.Println(resID)
	render.Template(w, r, "admin-reservations-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

func (m *Repository) AdminUpdateReservation(w http.ResponseWriter, r *http.Request) {
	resID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error parsing url param")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	// Get reservation from the database
	res, err := m.DBRepo.GetReservationByID(resID)
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error fetching reservation")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error parsing form")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	// Form validation
	form := forms.New(r.Form)
	form.Required("first_name", "last_name", "email")
	form.IsEmail("email")
	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "Form validation failure")
		w.Write([]byte("Form validation failure"))
		return
	}

	// Update the reservation in-memory
	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")

	// Write updated reservation to the database
	err = m.DBRepo.UpdateReservation(res)
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error updating reservation")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	src := chi.URLParam(r, "src")
	m.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

// AdminReservationsCalendar renders the admin reservations calendar page for the given year and month (default current year and month)
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// Assume that there is no month or year specified
	now := time.Now()
	if r.URL.Query().Get("y") != "" && r.URL.Query().Get("m") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}
	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	// Get the first and last day of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation) // I think this is unnecessary, it's the same as now
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["num_days_in_month"] = lastOfMonth.Day()

	data := make(map[string]interface{})
	data["now"] = now

	rooms, err := m.DBRepo.GetAllRooms()
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Internal error getting all rooms from the database")
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusInternalServerError)
		return
	}
	data["rooms"] = rooms
	// Focus on the current month for the current year.
	// For every room (we/owner only have two rooms), find all room restrictions within this month. Each room restriction
	// means the corresponding room is restricted either due to other customer reserves already or the owner blocks it.
	for _, rm := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-02")] = 0
		}

		// Get all room restrictions (reservation/blocked) for the currect room
		roomRestrictions, err := m.DBRepo.GetRoomRestrictionsForRoomByDate(rm.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			m.App.ErrorLog.Println(err)
			m.App.Session.Put(r.Context(), "error", "Internal error retrieving room restrictions for the given room")
			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusInternalServerError)
			return
		}

		for _, y := range roomRestrictions {
			if y.ReservationID > 0 { // restricted due to the room being reserved by other customers
				for d := y.StartDate; !d.After(y.EndDate); d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			} else { // restricted due to the room being blocked by the owner
				for d := y.StartDate; !d.After(y.EndDate); d = d.AddDate(0, 0, 1) {
					blockMap[d.Format("2006-01-2")] = y.RestrictionID
				}
			}
		}

		// One reservationMap for every room, and one blockMap for every room
		data[fmt.Sprintf("reservation_map_%d", rm.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", rm.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", rm.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	resID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error parsing url param")
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	src := chi.URLParam(r, "src")

	// Write to the database
	err = m.DBRepo.UpdateProcessedForReservation(resID, 1)
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Internal error marking reservation as processed")
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

// AdminDeleteReservation deletes the reservation from the database
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	resID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Error parsing url param")
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	src := chi.URLParam(r, "src")

	// Write to the database
	err = m.DBRepo.DeleteReservation(resID)
	if err != nil {
		m.App.ErrorLog.Println(err)
		m.App.Session.Put(r.Context(), "error", "Internal error deleting reservation "+string(resID))
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservation-%s", src), http.StatusSeeOther)
}
