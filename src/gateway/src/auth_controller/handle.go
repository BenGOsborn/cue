package auth_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/bengosborn/cue/utils"
)

var expiryTime = 5 * time.Minute
var length = 32

// Handle the authentication redirect
func HandleAuth(logger *log.Logger, client *utils.Redis, authenticator *utils.Authenticator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a new session
		key, err := utils.GenerateRandomString(length)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     utils.AuthCSRFCookie,
			Value:    key,
			Path:     "/",
			MaxAge:   int(expiryTime.Seconds()),
			HttpOnly: true,
		}

		http.SetCookie(w, &cookie)

		// Create a redirect URL
		state, err := utils.GenerateRandomString(length)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		redirectUrl := authenticator.GetAuthURL(state)

		if err := client.Set(utils.AuthCSRFCookie, key, state, expiryTime); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}
}
