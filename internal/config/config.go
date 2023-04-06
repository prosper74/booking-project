package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/atuprosper/booking-project/internal/models"
)

type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
	MailChannel   chan models.MailData
}
