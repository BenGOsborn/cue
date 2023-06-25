package gateway_controller

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/utils"
	utils "github.com/bengosborn/cue/utils"
)

// Attach the route to the server and start associated functions
func Attach(server *http.ServeMux, path string, connections *gwUtils.Connections, broker utils.Broker, lock *utils.ResourceLockDistributed, logger *log.Logger, process func(string, *gwUtils.Message) error) {
	server.HandleFunc(path, HandleWs(connections, logger, process))

	go ProcessMessages(connections, broker, lock, logger)
}
