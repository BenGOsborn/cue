package auth_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/bengosborn/cue/utils"
)

// Handle the authentication redirect
func Handle(logger *log.Logger, client *utils.Redis, authenticator *utils.Authenticator, expiryTime time.Duration) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a new session
		key, err := utils.GenerateRandomString(32)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     utils.AuthSession,
			Value:    key,
			Path:     "/",
			MaxAge:   int(expiryTime),
			HttpOnly: true,
		}

		http.SetCookie(w, &cookie)

		// Create a redirect URL
		state, err := utils.GenerateRandomString(32)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		redirectUrl := authenticator.GetAuthURL(state)

		client.Set(utils.AuthSession, key, state, expiryTime)
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}
}
