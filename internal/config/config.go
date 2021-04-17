package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/jackedelic/bookings/internal/models"
)

// AppConfig holds the application config
type AppConfig struct {
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	Session       *scs.SessionManager
	MailChan      chan models.MailData
	InProduction  bool
	UseCache      bool
	DBHost        string
	DBPort        int
	DBName        string
	DBUser        string
	DBPassword    string
	DBSSL         string
}
