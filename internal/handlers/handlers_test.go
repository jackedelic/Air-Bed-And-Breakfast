package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	{"search availability page", "/search-availability", "GET", []postData{}, http.StatusOK},
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
			t.Errorf("for %s, expected %d but got %d", test.name, test.expectedStatusCode, resp.StatusCode)
		}
	}
}
