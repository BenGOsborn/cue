package gateway

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

// Process incoming messages
func receive(id string, connections *gwUtils.Connections, logger *log.Logger) {
	for {
		// Read and process messages
		if ok, err := connections.Apply(id, func(id string, conn *websocket.Conn) error {
			_, p, err := conn.ReadMessage()

			if err != nil {
				return err
			}

			Process(gwUtils.Message{Id: id, Message: string(p)}, logger)

			return nil
		}); !ok || err != nil {
			logger.Println(err)
			connections.Remove(id)
			return
		}
	}
}

// Handle incoming connection
func Handle(connections *gwUtils.Connections, logger *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			logger.Println(err)
			return
		}

		// Add connection to connection pool
		id := uuid.NewString()
		connections.Add(id, conn)

		// Start receiving messages
		receive(id, connections, logger)
	}
}
