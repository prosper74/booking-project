package forms

import (
	"net/http"
	"net/url"
)

// A struct that holds all the information assciated with the form and embeds a url.Values object
type Form struct {
	url.Values
	Errors errors
}

// Valid, returns true if there are errors, otherwise return false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// New, initializes a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// HasField, checks if form field is in post and is not empty
func (f *Form) HasField(field string, r *http.Request) bool {
	x := r.Form.Get(field)
	if x == "" {
		f.Errors.Add(field, "This field is required")
		return false
	}
	return true
}
