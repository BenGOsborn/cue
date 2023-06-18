package gateway

import (
	"encoding/json"
	"fmt"
	"log"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	utils "github.com/bengosborn/cue/utils"
	"github.com/gorilla/websocket"
)

// Send a queued message
func Worker(connections *gwUtils.Connections, messages <-chan *utils.QueueMessage, logger *log.Logger) {
	for msg := range messages {
		// Send the message
		if ok, err := connections.Apply(msg.Receiver, func(id string, conn *websocket.Conn) error {
			data, err := json.Marshal(msg)

			if err != nil {
				return err
			}

			if err := conn.WriteMessage(1, data); err != nil {
				return err
			}

			logger.Println("Worker.sent: sent message to connection")

			return nil
		}); !ok || err != nil {
			if !ok {
				logger.Println("Worker.error: id does not exist")
			} else {
				logger.Println(fmt.Sprint("Worker.error: ", err))
			}
		}
	}
}
