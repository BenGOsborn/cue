package auth_controller

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/utils"
)

// Handle the authentication callback
func HandleLogout(session *gwUtils.Session, logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Delete the session
		sessionCookie, err := r.Cookie(gwUtils.SessionCookie)
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
			Name:     gwUtils.SessionCookie,
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
