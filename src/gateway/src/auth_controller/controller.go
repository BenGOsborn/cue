package auth_controller

import (
	"log"
	"net/http"

	"github.com/bengosborn/cue/utils"
)

// Attach the route to the server
func Attach(server *http.ServeMux, prefix string, logger *log.Logger, client *utils.Redis, authenticator *utils.Authenticator) {
	server.HandleFunc(prefix, HandleAuth(logger, client, authenticator))
	server.HandleFunc(prefix+"/callback", HandleCallback(client, authenticator, logger))
	server.HandleFunc(prefix+"/logout", HandleLogout(logger))
}
