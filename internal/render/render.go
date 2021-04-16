package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/jackedelic/bookings/internal/config"
	"github.com/jackedelic/bookings/internal/models"
	"github.com/justinas/nosurf"
)

var pathToTemplates = "./templates"

var functions = template.FuncMap{
	"humanDate":  HumanDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
}

// HumanDate returns time in "dd mmm yyyy" format
func HumanDate(t time.Time) string {
	return t.Format("02 Jan 2006")
}

func FormatDate(t time.Time, layout string) string {
	return t.Format(layout)
}

// Iterate returns a slice of int from 0 to count
func Iterate(count int) []int {
	var items []int
	for i := 1; i <= count; i++ {
		items = append(items, i)
	}
	return items
}

var app *config.AppConfig

// NewRenderer assigns the input AppConfig to local app
func NewRenderer(a *config.AppConfig) {
	app = a
}

// AddDefaultData adds more data (e.g CSRFToken) onto the input TemplateData, and returns the extended TemplateData
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = true
	}
	return td
}

// Template renders templates and writes to ResponseWriter
func Template(w http.ResponseWriter, r *http.Request, filename string, td *models.TemplateData) error {
	var tmplCache map[string]*template.Template
	if app.UseCache {
		tmplCache = app.TemplateCache
	} else {
		tmplCache, _ = CreateTemplateCache()
	}

	t, ok := tmplCache[filename]
	if !ok {
		log.Printf("%s does not exist. ", filename)
		return fmt.Errorf("%s does not exists", filename)
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	err := t.Execute(buf, td)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return err
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		fmt.Println("error writing template to the response writer", err)
		return err
	}
	return nil
}

// CreateTemplateCache creates a mapping of template file name to its parsed template.
func CreateTemplateCache() (map[string]*template.Template, error) {
	fmt.Println("create template cache")
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page) // last element of the path. e.g about.page.tmpl

		t, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// finds *.layout.tmpl files, just to check if there exists one.
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		// make sure there exists a *.layout.tmpl before we actually parse the template t
		if len(matches) > 0 {
			// parse this particular smtg.page.tmpl against all layout.tmpl files
			t, err = t.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates)) // same as t.ParseFiles("a.layout.tmpl","b.layout.tmpl"...)
			if err != nil {
				return myCache, err
			}
		}
		myCache[name] = t
	}
	return myCache, nil
}
