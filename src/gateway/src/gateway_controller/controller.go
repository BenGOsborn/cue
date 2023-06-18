package gateway_controller

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	utils "github.com/bengosborn/cue/utils"
)

// Attach the route to the server and start associated functions
func Attach(server *http.ServeMux, path string, connections *gwUtils.Connections, queues []*utils.Queue, logger *log.Logger, process func(string, *gwUtils.Message) error) {
	server.HandleFunc(path, HandleWs(connections, logger, process))

	for _, queue := range queues {
		go ProcessQueue(connections, queue, logger)
	}
}
