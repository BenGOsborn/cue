package auth_controller

import (
	"fmt"
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/utils"
)

// Attach the route to the server
func Attach(server *http.ServeMux, prefix string, logger *log.Logger, session *gwUtils.Session, authenticator *gwUtils.Authenticator) {
	server.HandleFunc(prefix, HandleAuth(logger, session, authenticator))
	server.HandleFunc(fmt.Sprint(prefix, "/callback"), HandleCallback(session, authenticator, logger))
	server.HandleFunc(fmt.Sprint(prefix, "/logout"), HandleLogout(session, logger))
}
