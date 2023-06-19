package auth_controller

import (
	"log"
	"net/http"

	"github.com/bengosborn/cue/utils"
)

// Handle the authentication callback
func HandleLogout(session *utils.Session, logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Delete the session
		sessionCookie, err := r.Cookie(utils.SessionCookie)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := session.Delete(sessionCookie.Value); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Remove the auth cookie
		authCookie := http.Cookie{
			Name:     utils.SessionCookie,
			Value:    "",
			Path:     "/",
			MaxAge:   int(-1),
			HttpOnly: true,
		}

		http.SetCookie(w, &authCookie)

		logger.Println("handlelogout.success: logged out")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
