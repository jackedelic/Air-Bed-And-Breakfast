package forms

import (
	"net/http"
	"net/url"
)

// Form creates a custom Form object, embeds a url.Values object
type Form struct {
	url.Values
	Errors errors
}

// New initializes a form struct
func New(data url.Values) *Form {
	return &Form{data, errors(map[string][]string{})}
}

// Has checks if form field is in POST request and not empty
func Has(field string, r *http.Request) bool {
	x := r.Form.Get(field)
	return x != ""
}
