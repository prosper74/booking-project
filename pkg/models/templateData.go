package models

// We create this seperate file to avoid Go 'import cycle not allowed error
// This is a struct that holds data sent from handlers to templates
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float32
	// We use type 'interface{}' whe the type is not known
	Data map[string]interface{}
	//Cross Site Request Forgery Token - a security token for our forms
	CSRFToken string
	Flash     string
	Warning   string
	Error     string
}
