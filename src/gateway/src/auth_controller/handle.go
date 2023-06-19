package auth_controller

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	"github.com/bengosborn/cue/helpers"
	"github.com/bengosborn/cue/utils"
)

// Handle the authentication redirect
func HandleAuth(logger *log.Logger, session *utils.Session, authenticator *utils.Authenticator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a new session
		sessionId, err := session.Create()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     utils.SessionCookie,
			Value:    sessionId,
			Path:     "/",
			MaxAge:   int(utils.SessionExpiry.Seconds()),
			HttpOnly: true,
		}

		http.SetCookie(w, &cookie)

		// Create a redirect URL
		state, err := helpers.GenerateRandomString(32)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		redirectUrl := authenticator.GetAuthURL(state)

		if err := session.Set(sessionId, gwUtils.SessionStateKey, state); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		logger.Println("handleauth.success: set csrf cookie")

		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}
}
