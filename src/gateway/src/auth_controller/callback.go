package auth_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/bengosborn/cue/utils"
)

// Handle the authentication callback
func HandleCallback(client *utils.Redis, authenticator *utils.Authenticator, logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify the CSRF token
		csrfCookie, err := r.Cookie(utils.AuthCSRFCookie)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		storedState := r.URL.Query().Get("state")

		if storedState == "" {
			http.Error(w, "Stored state not found", http.StatusInternalServerError)
			return
		}

		state := client.Get(utils.AuthCSRFCookie, csrfCookie.Value)

		if state != storedState {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		// Exchange the authorization code for a token
		code := r.URL.Query().Get("code")

		token, err := authenticator.ExchangeCodeForToken(code)

		if err != nil {
			http.Error(w, "Failed to exchange authorization code for token", http.StatusInternalServerError)
			return
		}

		// Set the access cookie
		authCookie := http.Cookie{
			Name:     utils.AuthAccessCookie,
			Value:    "Bearer " + token.AccessToken,
			Path:     "/",
			MaxAge:   int(time.Until(token.Expiry).Seconds()),
			HttpOnly: true,
		}

		http.SetCookie(w, &authCookie)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
