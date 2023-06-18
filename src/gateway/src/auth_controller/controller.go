package auth_controller

import (
	"log"
	"net/http"

	"github.com/bengosborn/cue/utils"
)

// Attach the route to the server
func Attach(server *http.ServeMux, path string, callbackPath string, logger *log.Logger, client *utils.Redis, authenticator *utils.Authenticator) {
	server.HandleFunc(path, HandleAuth(logger, client, authenticator))
	server.HandleFunc(callbackPath, HandleCallback(client, authenticator, logger))
}
