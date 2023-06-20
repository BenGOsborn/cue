package gateway_controller

import (
	"encoding/json"
	"fmt"
	"log"

	gwUtils "github.com/bengosborn/cue/gateway/utils"
	utils "github.com/bengosborn/cue/utils"
	"github.com/gorilla/websocket"
)

// Process queued messages
func ProcessQueue(connections *gwUtils.Connections, queue *utils.Queue, logger *log.Logger) {
	if err := queue.Listen(func(queueMessage *utils.QueueMessage) bool {
		if ok, err := connections.Apply(queueMessage.Receiver, func(_ string, conn *websocket.Conn) error {
			data, err := json.Marshal(queueMessage)

			if err != nil {
				return err
			}

			if err := conn.WriteMessage(1, data); err != nil {
				return err
			}

			logger.Println("processqueue.success: sent message to connection")

			return nil
		}); !ok || err != nil {
			if !ok {
				logger.Println("processqueue.error: id does not exist")
			} else {
				logger.Println(fmt.Sprint("processqueue.error: ", err))
			}

			return false
		}

		return true
	}, nil); err != nil {
		logger.Fatalln(fmt.Sprint("processqueue.error: ", err))
	}
}
