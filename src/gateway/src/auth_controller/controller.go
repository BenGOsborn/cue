package auth_controller

import (
	"log"
	"net/http"
)

// Attach the route to the server
func Attach(server *http.ServeMux, path string, callbackPath string, logger *log.Logger) {
	// server.HandleFunc(path, HandleWs(connections, logger, process))
}
