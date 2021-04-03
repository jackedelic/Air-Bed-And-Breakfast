package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	// {"search availability json", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "31-01-2021"},
	// 	{key: "end", value: "01-02-2021"},
	// 	{key: "room_id", value: "1"},
	// }, http.StatusOK},
	// {"make reservation", "/make-reservation", "POST", []postData{
	// 	{key: "first_name", value: "Jack"},
	// 	{key: "last_name", value: "Wong"},
	// 	{key: "email", value: "jackwong3101@yahoo.com"},
	// }, http.StatusOK},
	// {"search available rooms", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "31-01-2021"},
	// 	{key: "end", value: "01-02-2021"},
	// }, http.StatusOK},
	// {"receive json", "/receive-json", "POST", []postData{
	// 	{key: "", value: ""},
	// }, http.StatusOK},
}

func aTestHandlers(t *testing.T) {
	routes := getRoutes()
	server := httptest.NewTLSServer(routes)
	defer server.Close()

	for _, test := range theTests {
		client := server.Client()
		switch test.method {
		case "GET":
			resp, err := client.Get(server.URL + test.urlPath)
			if err != nil {
				t.Log(err)
				t.Error(err)
			}

			if resp.StatusCode != test.expectedStatusCode {
				t.Errorf("for %s: %s, expected %d but got %d", test.method, test.name, test.expectedStatusCode, resp.StatusCode)
			}
			// case "POST":
			// 	var formData = url.Values{}
			// 	for _, pData := range test.params {
			// 		formData.Add(pData.key, pData.value)
			// 	}

			// 	resp, err := client.PostForm(server.URL+test.urlPath, formData)
			// 	if err != nil {
			// 		t.Log(err)
			// 		t.Error(err)
			// 	}

			// 	if resp.StatusCode != test.expectedStatusCode {
			// 		t.Errorf("for %s: %s, expected %d but got %d", test.method, test.name, test.expectedStatusCode, resp.StatusCode)
			// 	}
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
		body               string
		reservation        interface{} // allows for nil
		expectedStatusCode int
	}{
		// Test for valid post data
		{
			description: "Test for valid post data",
			body: "first_name=Jordan" + "&" + "last_name=Peele" + "&" +
				"email=jordan@comedycentral.com" + "&" + "room_id=1",
			reservation: models.Reservation{
				StartDate: time.Now(),
				EndDate:   time.Now(),
			},
			expectedStatusCode: http.StatusSeeOther,
		},
		// Test for missing body
		{
			description:        "Test for missing body",
			body:               "",
			reservation:        models.Reservation{},
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		// Test for missing Reservation in the session
		{
			description: "Test for missing Reservation in the session",
			body: "first_name=Jordan" + "&" + "last_name=Peele" + "&" +
				"email=jordan@comedycentral.com" + "&" + "room_id=1",
			reservation:        nil,
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		// Test for room_id not being integer
		{
			description: "Test for room_id not being integer",
			body: "first_name=Jordan" + "&" + "last_name=Peele" + "&" +
				"email=jordan@comedycentral.com" + "&" + "room_id=notinteger",
			reservation: models.Reservation{
				StartDate: time.Now(),
				EndDate:   time.Now(),
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		// Test for invalid form (first_name required)
		{
			description: "Test for invalid form (first_name required)",
			body: "last_name=Peele" + "&" +
				"email=jordan@comedycentral.com" + "&" + "room_id=1",
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
			body: "first_name=Jordan" + "&" + "last_name=Peele" + "&" +
				"email=jordan@comedycentral.com" + "&" + "room_id=2",
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
			body: "first_name=Jack" + "&" + "last_name=Peele" + "&" +
				"email=jordan@comedycentral.com" + "&" + "room_id=1000",
			reservation: models.Reservation{
				StartDate: time.Now(),
				EndDate:   time.Now(),
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
	}

	// Start testing each test data
	for i := 0; i < len(testData); i++ {
		if testData[i].body == "" {
			req, _ = http.NewRequest("POST", "/make-reservation", nil) // no session yet
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(testData[i].body)) // no session yet
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

// Gets the X-Session from request header and load into a context.Context object.
func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		Repo.App.ErrorLog.Println(err)
	}
	return ctx
}
