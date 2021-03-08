package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/jackedelic/go-overview-trevor-sawler/pkg/config"
	"github.com/jackedelic/go-overview-trevor-sawler/pkg/models"
)

var functions = template.FuncMap{}

var app *config.AppConfig

// NewConfig assigns the input AppConfig to local app
func NewConfig(a *config.AppConfig) {
	app = a
}

// AddDefaultData adds more data onto the input TemplateData, and returns the extended TemplateData
func AddDefaultData(td *models.TemplateData) *models.TemplateData {
	return td
}

// RenderTemplate renders templates
func RenderTemplate(w http.ResponseWriter, filename string, td *models.TemplateData) {
	var tmplCache map[string]*template.Template
	if app.UseCache {
		tmplCache = app.TemplateCache
	} else {
		tmplCache, _ = CreateTemplateCache()
	}

	t, ok := tmplCache[filename]
	if !ok {
		log.Fatal(fmt.Sprintf("%s does not exist. ", filename))
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td)

	err := t.Execute(buf, td)
	if err != nil {
		fmt.Println("Error parsing template:", err)
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		fmt.Println("error writing template to the response writer", err)
	}
}

// RenderTemplates creates a mapping of template file name to its parsed template.
func CreateTemplateCache() (map[string]*template.Template, error) {
	fmt.Println("create template cache")
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob("./templates/*.page.tmpl")
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
		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			return myCache, err
		}

		// make sure there exists a *.layout.tmpl before we actually parse the template t
		if len(matches) > 0 {
			// parse this particular smtg.page.tmpl against all layout.tmpl files
			t, err = t.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				return myCache, err
			}
		}
		myCache[name] = t
	}
	return myCache, nil
}
