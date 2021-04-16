package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jackedelic/bookings/internal/models"
)

// postData holds the key-value pairs of form inputs (name-value pairs)
type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	urlPath            string
	method             string
	expectedStatusCode int
}{
	{"home page", "/", "GET", http.StatusOK},
	{"about page", "/about", "GET", http.StatusOK},
	{"make reservation page", "/make-reservation", "GET", http.StatusOK},
	{"reservation summary page", "/reservation-summary", "GET", http.StatusOK},
	{"generals quarters page", "/generals-quarters", "GET", http.StatusOK},
	{"majors suite page", "/majors-suite", "GET", http.StatusOK},
	{"contact page", "/contact", "GET", http.StatusOK},
	{"non-existent route", "/havefun/burgerking", "GET", http.StatusNotFound},
	{"login page", "/user/login", "GET", http.StatusOK},
	{"admin dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"admin show new reservations", "/admin/reservations-new", "GET", http.StatusOK},
	{"admin show all reservations", "/admin/reservations-all", "GET", http.StatusOK},
	{"admin show reservations calendar", "/admin/reservations-calendar", "GET", http.StatusOK},
	{"admin show reservation", "/admin/reservations/new/1", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	server := httptest.NewTLSServer(routes)
	defer server.Close()

	for _, test := range theTests {
		client := server.Client()
		resp, err := client.Get(server.URL + test.urlPath)
		if err != nil {
			t.Log(err)
			t.Error(err)
		}

		if resp.StatusCode != test.expectedStatusCode {
			t.Errorf("for %s: %s, expected %d but got %d", test.method, test.name, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

// Tests GET /make-reservation endpoint.
// Create models.Reservation and a new session (stored in context.Context).
// Then puts the reservation into the session (remember the session is inside the context).
// Then puts the context (containing the session) into the request (created using http.NewRequest())
func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1, RoomName: "General's Quarter",
		},
	}
	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)         // context containing the session
	req = req.WithContext(ctx) // request containing this context containing the session.

	resRecorder := httptest.NewRecorder()
	// session is global variable from setup_test.go, assigned to app.Session
	session.Put(ctx, "reservation", reservation) //  the server will remember the session token (in ctx)

	handler := http.HandlerFunc(Repo.MakeReservation) // notice we don need getRoutes. Here we're building our handler ourselves
	handler.ServeHTTP(resRecorder, req)
	if resRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", resRecorder.Code, http.StatusOK)
	}

	// test case where Reservation is NOT in session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	resRecorder = httptest.NewRecorder()
	handler.ServeHTTP(resRecorder, req)
	if resRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", resRecorder.Code, http.StatusTemporaryRedirect)
	}

	// test with non-existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	resRecorder = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(resRecorder, req)
	if resRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", resRecorder.Code, http.StatusOK)
	}
}

// TestRepository_PostReservation tests the PostReservation handler with Repository obj as the receiver.
// It tests POST /make-reservation endpoint.
func TestRepository_PostReservation(t *testing.T) {
	var (
		ctx         context.Context
		req         *http.Request
		resRecorder *httptest.ResponseRecorder
		handler     http.HandlerFunc
	)
	var testData = []struct {
		description        string
		body               map[string][]string
		reservation        interface{} // allows for nil
		expectedStatusCode int
	}{
		// Test for valid post data
		{
			description: "Test for valid post data",
			body: map[string][]string{
				"first_name": {"Jordan"},
				"last_name":  {"Peele"},
				"email":      {"jordan@comedycentral.com"},
				"room_id":    {"1"}},
			reservation: models.Reservation{
				StartDate: time.Now(),
				EndDate:   time.Now(),
			},
			expectedStatusCode: http.StatusSeeOther,
		},
		// Test for missing body
		{
			description:        "Test for missing body",
			body:               map[string][]string{},
			reservation:        models.Reservation{},
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		// Test for missing Reservation in the session
		{
			description: "Test for missing Reservation in the session",
			body: map[string][]string{
				"first_name": {"Jordan"},
				"last_name":  {"Peele"},
				"email":      {"jordan@comedycentral.com"},
				"room_id":    {"1"}},
			reservation:        nil,
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		// Test for room_id not being integer
		{
			description: "Test for room_id not being integer",
			body: map[string][]string{
				"first_name": {"Jordan"},
				"last_name":  {"Peele"},
				"email":      {"jordan@comedycentral.com"},
				"room_id":    {"notinteger"}},
			reservation: models.Reservation{
				StartDate: time.Now(),
				EndDate:   time.Now(),
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		// Test for invalid form (first_name required)
		{
			description: "Test for invalid form (first_name required)",
			body: map[string][]string{
				"last_name": {"Peele"},
				"email":     {"jordan@comedycentral.com"},
				"room_id":   {"1"}},
			reservation: models.Reservation{
				StartDate: time.Now(),
				EndDate:   time.Now(),
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		// Test for error inserting into reservations table (Repo.InsertReservation)
		// Repo.InsertReservation returns error for room_id of 2
		{
			description: "Test for error inserting into reservation table",
			body: map[string][]string{
				"first_name": {"Jordan"},
				"last_name":  {"Peele"},
				"email":      {"jordan@comedycentral.com"},
				"room_id":    {"2"}},
			reservation: models.Reservation{
				StartDate: time.Now(),
				EndDate:   time.Now(),
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		// Test for error inserting into room_restrictions table (Repo.InsertRoomRestriction)
		// Repo.InsertRoomRestriction returns error for room_id of 1000
		{
			description: "Test for error inserting into room_restrictions table",
			body: map[string][]string{
				"first_name": {"Jack"},
				"last_name":  {"Peele"},
				"email":      {"jordan@comedycentral.com"},
				"room_id":    {"1000"}},
			reservation: models.Reservation{
				StartDate: time.Now(),
				EndDate:   time.Now(),
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
	}

	// Start testing each test data
	for i := 0; i < len(testData); i++ {
		// Create url encoded form string
		formString := url.Values(testData[i].body).Encode()
		if len(testData[i].body) == 0 {
			req, _ = http.NewRequest("POST", "/make-reservation", nil) // no session yet
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(formString)) // no session yet
		}

		ctx = getCtx(req)          // get X-Session token and create a context with the session token
		req = req.WithContext(ctx) // now the request has session
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		session.Put(req.Context(), "reservation", testData[i].reservation)

		resRecorder = httptest.NewRecorder() // satisfies the requirement of being a ResponseWriter

		handler = http.HandlerFunc(Repo.PostReservation)

		handler.ServeHTTP(resRecorder, req)
		if resRecorder.Code != testData[i].expectedStatusCode {
			t.Errorf("Test description: %s \nPostReservation handlers returned the wrong response code: got %d, wanted %d", testData[i].description, resRecorder.Code, testData[i].expectedStatusCode)
		}
	}
}

func TestRepository_SearchAvailabilityJSON(t *testing.T) {
	var (
		ctx         context.Context
		req         *http.Request
		resRecorder *httptest.ResponseRecorder
		handler     http.HandlerFunc
	)
	testData := []struct {
		description          string
		body                 map[string][]string
		expectedJSONResponse JSONResponse
	}{
		{
			description:          "empty form body",
			body:                 map[string][]string{},
			expectedJSONResponse: JSONResponse{Ok: false, Message: "Internal server error parsing form"},
		},
		{
			description: "room_id is not an integer",
			body: map[string][]string{
				"start":   {"01-01-2050"},
				"end":     {"01-01-2050"},
				"room_id": {"invalid"}},
			expectedJSONResponse: JSONResponse{Ok: false, Message: "Internal server error processing room_id"},
		},
		{
			description: "missing start_date in the form",
			body: map[string][]string{
				"end":     {"01-01-2050"},
				"room_id": {"1"}},
			expectedJSONResponse: JSONResponse{Ok: false, Message: "Internal server error retrieving start_date from form"},
		},
		{
			description: "missing end_date in the form",
			body: map[string][]string{
				"start":   {"01-01-2050"},
				"room_id": {"1"}},
			expectedJSONResponse: JSONResponse{Ok: false, Message: "Internal server error parsing end_date from form"},
		},
		{
			description: "error SearchAvailabilityByDatesByRoomID",
			body: map[string][]string{
				"start":   {"01-01-2050"},
				"end":     {"01-01-2050"},
				"room_id": {"1"}},
			expectedJSONResponse: JSONResponse{Ok: false, Message: "Internal server error searching availability by dates by room id"},
		},
	}

	for i := 0; i < len(testData); i++ {
		// Encode body
		formString := url.Values(testData[i].body).Encode()
		// Create request
		if len(testData[i].body) == 0 {
			req, _ = http.NewRequest("POST", "/search-availability-json", nil)
		} else {
			req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(formString))
		}

		// Set request header on content-type
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Create context with session in the request
		ctx = getCtx(req)

		// Put the context with session into the request
		req = req.WithContext(ctx)

		// Make ResponseRecorder
		resRecorder = httptest.NewRecorder()

		// Create SearchAvailabilityJSON handler
		handler = http.HandlerFunc(Repo.SearchAvailabilityJSON)

		// Makes request to out handler
		handler.ServeHTTP(resRecorder, req)

		// Receives response and processes the bytes into JSONResponse struct
		var jsonResponse JSONResponse
		err := json.Unmarshal(resRecorder.Body.Bytes(), &jsonResponse)
		if err != nil {
			t.Error("failed to parse json")
		}
		desc := testData[i].description
		expectedOk := testData[i].expectedJSONResponse.Ok
		expectedMsg := testData[i].expectedJSONResponse.Message
		if jsonResponse.Ok != expectedOk {
			t.Errorf("Test description: %s \nThe json response had the wrong Ok value: got: %t, wanted: %t",
				desc, jsonResponse.Ok, expectedOk)
		}
		if jsonResponse.Message != expectedMsg {
			t.Errorf("Test description: %s \nThe json response had the wrong Message value: got %s, wanted %s",
				desc, jsonResponse.Message, expectedMsg)
		}
	}
}

// Tests POST /search-availability
var testPostAvailabilityData = []struct {
	name               string
	postedData         url.Values
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "rooms not available :(",
		postedData: url.Values{
			"start": {"01-01-2050"}, // testingDBRepo.SearchAvailableRoomsByDate returns empty slice for "01-01-2050" (start or end date)
			"end":   {"02-01-2050"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "rooms are available :)",
		postedData: url.Values{
			"start": {"01-01-2040"},
			"end":   {"02-01-2040"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "empty post body",
		postedData:         url.Values{},
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "start date wrong format",
		postedData: url.Values{
			"start": {"2040-02-01"},
			"end":   {"01-02-2040"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "end date wrong format",
		postedData: url.Values{
			"start": {"01-02-2040"},
			"end":   {"2040-02-01"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
}

func TestRepository_PostSearchAvailability(t *testing.T) {
	for _, e := range testPostAvailabilityData {
		req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(e.postedData.Encode()))

		// get the context with session
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		// set the request header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// make our PostSearchAvailability handler an http.HandlerFunc and call
		handler := http.HandlerFunc(Repo.PostSearchAvailability)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s gave wrong status code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{"valid-credentials", "me@here.ca", http.StatusSeeOther, "", "/"}, // testingDBRepo.Authenticate recognizes me@here.ca as the only valid emailk
	{"invalid-credentials", "jack@nimble.com", http.StatusSeeOther, "", "/user/login"},
	{"invalid-data", "invalid-email@", http.StatusUnauthorized, `action="/user/login"`, ""},
}

func TestLogin(t *testing.T) {
	// range through all tests
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "password")

		// create a request
		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.PostLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		// check redirected url if the response redirects
		if e.expectedLocation != "" {
			// get the url from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		// check for expected html contained in the received HTML
		if e.expectedHTML != "" {
			// read the response body into a string
			actualHTML := rr.Body.String()
			if !strings.Contains(actualHTML, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s, but did not", e.name, e.expectedHTML)
			}

		}
	}
}

// Gets the X-Session from request header and load into a context.Context object.
func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		Repo.App.ErrorLog.Println(err)
	}
	return ctx
}
