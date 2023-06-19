package auth_controller

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/utils"
)

// Handle the authentication callback
func HandleCallback(session *gwUtils.Session, authenticator *gwUtils.Authenticator, logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the session
		sessionCookie, err := r.Cookie(gwUtils.SessionCookie)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		sessionData, err := session.Get(sessionCookie.Value)
		if err != nil {
			http.Error(w, "Invalid session cookie", http.StatusBadRequest)
			return
		}

		// Verify the CSRF token
		state := r.URL.Query().Get("state")

		if state == "" || sessionData.CSRFToken != state {
			http.Error(w, "Invalid stored state", http.StatusInternalServerError)
			return
		}

		// Store the id token
		code := r.URL.Query().Get("code")

		rawIdToken, _, err := authenticator.ExchangeCodeForToken(code)
		if err != nil {
			http.Error(w, "Failed to exchange authorization code for token", http.StatusInternalServerError)
			return
		}

		session.Set(sessionCookie.Value, &gwUtils.SessionData{Token: rawIdToken})

		logger.Println("handlecallback.success: authenticated new session")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
