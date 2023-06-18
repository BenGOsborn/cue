package auth_controller

import (
	"log"
	"net/http"

	"github.com/bengosborn/cue/utils"
)

// Handle the authentication redirect
func HandleAuth(logger *log.Logger, redis *utils.Redis, authenticator *utils.Authenticator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a new session
		key, err := utils.GenerateRandomString(utils.TokenLength)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     utils.AuthCSRFCookie,
			Value:    key,
			Path:     "/",
			MaxAge:   int(utils.TokenExpiryTime.Seconds()),
			HttpOnly: true,
		}

		http.SetCookie(w, &cookie)

		// Create a redirect URL
		state, err := utils.GenerateRandomString(utils.TokenLength)

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		redirectUrl := authenticator.GetAuthURL(state)

		if err := redis.Set(utils.AuthCSRFCookie, key, state, utils.TokenExpiryTime); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		logger.Println("handleauth.success: set csrf cookie")

		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}
}
