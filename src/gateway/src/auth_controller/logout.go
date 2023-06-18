package auth_controller

import (
	"log"
	"net/http"

	"github.com/bengosborn/cue/utils"
)

// Handle the authentication callback
func HandleLogout(logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Remove the access cookie
		authCookie := http.Cookie{
			Name:     utils.AuthAccessCookie,
			Value:    "",
			Path:     "/",
			MaxAge:   int(-1),
			HttpOnly: true,
		}

		http.SetCookie(w, &authCookie)

		logger.Println("handlelogout.success: removed authentication cookie")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}