package render

import (
	"net/http"
	"testing"

	"github.com/jackedelic/bookings/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	td := &models.TemplateData{}
	req, err := http.NewRequest("GET", "/dummy-path", nil)
	if err != nil {
		t.Error(err)
	}
	req, err = withSession(req)
	session.Put(req.Context(), "flash", "123")

	if err != nil {
		t.Error(err)
	}

	td = AddDefaultData(td, req)
	if td.Flash != "123" {
		t.Error("flash value of 123 not found in session")
	}
}

func withSession(r *http.Request) (*http.Request, error) {
	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)
	return r, nil
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplates = "../../templates"

	// Creates template cache (necessary cur CreateTemplateCache requires initialization of cache)
	cache, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
	app.TemplateCache = cache

	// Creates a request
	req, err := http.NewRequest("GET", "/dummy-path", nil)
	if err != nil {
		t.Error(err)
	}

	// Add session to the newly created request
	req, err = withSession(req)
	if err != nil {
		t.Error(err)
	}

	// Test RenderTemplate using the created request
	var myWriter myResponseWriter
	err = RenderTemplate(&myWriter, req, "home.page.tmpl", &models.TemplateData{})
	if err != nil {
		t.Error("error writing template to browser")
	}

	err = RenderTemplate(&myWriter, req, "non-existent.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("rendered template that does not exist")
	}
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "../../templates"

	// Creates template cache (not necessary ?)
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}
