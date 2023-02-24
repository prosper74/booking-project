package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/atuprosper/booking-project/internal/config"
	"github.com/atuprosper/booking-project/internal/models"
	"github.com/justinas/nosurf"
)

var functions = template.FuncMap{}

var app *config.AppConfig
var pathToTemplates = "./templates"

// NewTemplates sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	// PopString puts data into our session until a new page is loaded
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")

	// Generate CSRFToken from nosurf
	td.CSRFToken = nosurf.Token(r)
	return td
}

func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {

	var tc map[string]*template.Template

	// Get the template cache from the app config
	// Check if we are in dev mode, load cache from disk, else load it from the template
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	// t holds the actual template found, while "ok" will return true if the template exists in our directory. If we get passed this, then we have the actual template that we want to use
	t, ok := tc[tmpl]
	if !ok {
		return errors.New("could not get template from cache")
	}

	// Create a bytesBuffer that will hold the information of the parsed template in memory, and put them in a byte
	buf := new(bytes.Buffer)

	// Before we execute the buffer, we want to attach the AddDefaultData
	td = AddDefaultData(td, r)

	//Execeute the tamplate file and put it in the buffer
	_ = t.Execute(buf, td)

	// Write the buffer to the resposeWriter(browser)
	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser", err)
		return err
	}

	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	// A [map] data structure to look up things very quickly. myCache is a cache that will hold all the templates
	myCache := map[string]*template.Template{}

	// get all the pages in the "templates directory" that ends with page.html
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.html", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// Loop through all the pages to get individual page and the filepath of the page\
	// The underscore _ means we are ignoring the index of the list
	for _, page := range pages {
		name := filepath.Base(page)
		fmt.Println("Page is currently", page)

		// Create a template set (ts), that will have functions "Funcs(functions), which are external functions not build into Go language"
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// Look for any files that ends with layout.html in the templates directory
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		// If we find any layout.html file, we want to pass them to our template set (ts)
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		// Add the template set(ts) we just created to our cache
		myCache[name] = ts
	}

	// Return myCache and ignore the value for error using nil. We have already dealt with all the posible errors
	return myCache, nil
}
