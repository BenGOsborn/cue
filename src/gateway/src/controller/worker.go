package gateway

import (
	"log"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	"github.com/gorilla/websocket"
)

// Send a queued message
func Worker(connections *gwUtils.Connections, messages <-chan gwUtils.Message, logger *log.Logger) {
	for msg := range messages {
		// Send the message
		connections.Apply(msg.Id, func(id string, conn *websocket.Conn) error {
			if err := conn.WriteMessage(1, []byte(id+": "+msg.Message)); err != nil {
				logger.Println(err)
			}

			return nil
		})
	}

}
