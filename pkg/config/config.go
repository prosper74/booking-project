package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
)

// This file should be made accessible by other packages, every package in this project can import it and use it. But this file should not import any package, it will only use the standard built-in Go packages. This will help us avoid an error called "import cycle"

// AppConfig holds the application configurations
// InfoLog allows you to create a log file and store informations in it
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	InPrduction   bool
	Session       *scs.SessionManager
}
