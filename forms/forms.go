package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Valid returns true if there are no errors. Otherwise false.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Form creates a custom Form object, embeds a url.Values object
type Form struct {
	url.Values // This is called embedded field in Go struct. All methods that can be called on url.Values (e.g Get)
	// can be called directly on a Form object.
	Errors errors
}

// New initializes a form struct
func New(data url.Values) *Form {
	return &Form{data, errors(map[string][]string{})}
}

// Has checks if form field is in POST request and not empty
func (f *Form) Has(field string) bool {
	x := f.Get(field)
	return x != ""
}

// Required populates f.Errors[field] if f does not have any of the fields
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// MinLength checks for field string minimum length
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long.", length))
		return false
	}
	return true
}

// IsEmail checks for valid email address
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}
