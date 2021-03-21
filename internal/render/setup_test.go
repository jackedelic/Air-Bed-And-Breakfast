package render

import (
	"encoding/gob"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/jackedelic/bookings/internal/config"
	"github.com/jackedelic/bookings/internal/models"
)

var session *scs.Session
var testApp config.AppConfig

func TestMain(m *testing.M) {
	gob.Register(models.Reservation{})
	// Register session for all requests
	testApp.InProduction = false
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = testApp.InProduction

	testApp.Session = session
	NewConfig(&testApp)
	os.Exit(m.Run())
}

// Mock ResponseWriter
type myResponseWriter struct {
}

func (w *myResponseWriter) Header() http.Header {
	var h http.Header
	return h
}

func (w *myResponseWriter) Write(b []byte) (int, error) {
	var i int = len(b)
	return i, nil
}

func (w *myResponseWriter) WriteHeader(statusCode int) {

}
