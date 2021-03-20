package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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
	{"search availability page", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"make reservation", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "Jack"},
		{key: "last_name", value: "Wong"},
		{key: "email", value: "jackwong3101@yahoo.com"},
	}, http.StatusOK},
	{"search availability", "/search-availability", "POST", []postData{
		{key: "start", value: "31-01-2021"},
		{key: "end", value: "01-02-2021"},
	}, http.StatusOK},
	{"receive json", "/receive-json", "POST", []postData{
		{key: "", value: ""},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
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
				t.Errorf("for %s, expected %d but got %d", test.name, test.expectedStatusCode, resp.StatusCode)
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
				t.Errorf("for %s, expected %d but got %d", test.name, test.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}
