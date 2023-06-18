package gateway

import (
	"encoding/json"
	"log"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	"github.com/gorilla/websocket"
)

// Send a queued message
func Worker(connections *gwUtils.Connections, messages <-chan *gwUtils.QueueMessage, logger *log.Logger) {
	for msg := range messages {
		// Send the message
		if ok, err := connections.Apply(msg.Id, func(id string, conn *websocket.Conn) error {
			data, err := json.Marshal(msg)

			if err != nil {
				return err
			}

			if err := conn.WriteMessage(1, data); err != nil {
				return err
			}

			return nil
		}); !ok || err != nil {
			if !ok {
				logger.Println("could not apply to given id")
			} else {
				logger.Println(err)
			}
		}
	}
}
