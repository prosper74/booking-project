package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/atuprosper/booking-project/internal/config"
)

var app *config.AppConfig

// Setup app config for new helpers
func NewHelpers(helper *config.AppConfig) {
	app = helper
}

func ClientError(responseWriter http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(responseWriter, http.StatusText(status), status)
}

func ServerError(responseWriter http.ResponseWriter, err error) {
	// err.Error() prints the nature of the error. debug.Stack() prints the detailed info about the nature of the error
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)

	// Send feedback to the user
	http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Check if a user is exists
func IsAuthenticated(request *http.Request) bool {
	exists := app.Session.Exists(request.Context(), "user_id")
	return exists
}
