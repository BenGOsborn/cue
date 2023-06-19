package auth_controller

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/utils"
	"github.com/bengosborn/cue/utils"
)

// Attach the route to the server
func Attach(server *http.ServeMux, prefix string, logger *log.Logger, session *utils.Session, authenticator *gwUtils.Authenticator) {
	server.HandleFunc(prefix, HandleAuth(logger, session, authenticator))
	server.HandleFunc(prefix+"/callback", HandleCallback(session, authenticator, logger))
	server.HandleFunc(prefix+"/logout", HandleLogout(session, logger))
}
