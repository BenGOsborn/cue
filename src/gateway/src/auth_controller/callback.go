package auth_controller

import (
	"log"
	"net/http"
	"time"

	"github.com/bengosborn/cue/utils"
)

// Handle the authentication callback
func HandleCallback(redis *utils.Redis, authenticator *utils.Authenticator, logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
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

		state := redis.Get(utils.AuthCSRFCookie, csrfCookie.Value)

		if state != storedState {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		// Exchange the authorization code for a token
		code := r.URL.Query().Get("code")

		rawIdToken, idToken, err := authenticator.ExchangeCodeForToken(code)
		if err != nil {
			http.Error(w, "Failed to exchange authorization code for token", http.StatusInternalServerError)
			return
		}

		// Set the auth cookie
		authCookie := http.Cookie{
			Name:     utils.AuthIdCookie,
			Value:    "Bearer " + rawIdToken,
			Path:     "/",
			MaxAge:   int(time.Until(idToken.Expiry).Seconds()),
			HttpOnly: true,
		}

		http.SetCookie(w, &authCookie)

		logger.Println("handlecallback.success: set authentication cookie")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
