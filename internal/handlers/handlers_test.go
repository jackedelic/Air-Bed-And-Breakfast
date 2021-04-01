package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
	params             []postData
	expectedStatusCode int
}{
	{"home page", "/", "GET", []postData{}, http.StatusOK},
	{"about page", "/about", "GET", []postData{}, http.StatusOK},
	{"make reservation page", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"reservation summary page", "/reservation-summary", "GET", []postData{}, http.StatusOK},
	{"generals quarters page", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"majors suite page", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"contact page", "/contact", "GET", []postData{}, http.StatusOK},
	{"search availability json", "/search-availability-json", "POST", []postData{
		{key: "start", value: "31-01-2021"},
		{key: "end", value: "01-02-2021"},
		{key: "room_id", value: "1"},
	}, http.StatusOK},
	{"make reservation", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "Jack"},
		{key: "last_name", value: "Wong"},
		{key: "email", value: "jackwong3101@yahoo.com"},
	}, http.StatusOK},
	{"search available rooms", "/search-availability", "POST", []postData{
		{key: "start", value: "31-01-2021"},
		{key: "end", value: "01-02-2021"},
	}, http.StatusOK},
	{"receive json", "/receive-json", "POST", []postData{
		{key: "", value: ""},
	}, http.StatusOK},
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
		case "POST":
			var formData = url.Values{}
			for _, pData := range test.params {
				formData.Add(pData.key, pData.value)
			}

			resp, err := client.PostForm(server.URL+test.urlPath, formData)
			if err != nil {
				t.Log(err)
				t.Error(err)
			}

			if resp.StatusCode != test.expectedStatusCode {
				t.Errorf("for %s: %s, expected %d but got %d", test.method, test.name, test.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}

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
	session.Put(ctx, "reservation", reservation) // If middleware uses session, then the server will remember the session token (in ctx)

	handler := http.HandlerFunc(Repo.MakeReservation) // notice we don need getRoutes. Here we're building our handler ourselves
	handler.ServeHTTP(resRecorder, req)
	if resRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", resRecorder.Code, http.StatusOK)
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		Repo.App.ErrorLog.Println(err)
	}
	return ctx
}
