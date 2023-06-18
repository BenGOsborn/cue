package gateway_controller

import (
	"fmt"
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

// Process incoming messages
func receive(id string, connections *gwUtils.Connections, logger *log.Logger, process func(string, *gwUtils.Message) error) {
	for {
		// Read and process messages
		if ok, err := connections.Apply(id, func(id string, conn *websocket.Conn) error {
			var message gwUtils.Message

			if err := conn.ReadJSON(&message); err != nil {
				return err
			}

			if err := process(id, &message); err != nil {
				return err
			}

			return nil
		}); !ok || err != nil {
			if !ok {
				logger.Println("receive.error: id does not exist")
			} else {
				logger.Println(fmt.Sprint("receive.error: ", err))
			}

			connections.Remove(id)

			logger.Println("receive.removed: removed connection")

			return
		}
	}
}

// Handle incoming connection
func HandleWs(connections *gwUtils.Connections, logger *log.Logger, process func(string, *gwUtils.Message) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			logger.Println(fmt.Sprint("handle.error: ", err))
			return
		}

		// Add connection to connection pool
		id := uuid.NewString()
		connections.Add(id, conn)

		logger.Println("handlews.connection: added new connection")

		// Start receiving messages
		receive(id, connections, logger, process)
	}
}
