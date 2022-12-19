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

// New, initializes a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// HasField, checks if form field is in post and is not empty
func (f *Form) HasField(field string, req *http.Request) bool {
	x := req.Form.Get(field)
	return x != ""
	// if x == "" {
	// return false
	// }
	// return true
}
