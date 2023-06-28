package gateway_controller

import (
	"encoding/json"
	"log"

	gwUtils "github.com/bengosborn/cue/gateway/utils"
	utils "github.com/bengosborn/cue/utils"
	"github.com/gorilla/websocket"
)

// Process messages from broker
func ProcessMessages(connections *gwUtils.Connections, broker utils.Broker, lock *utils.ResourceLockDistributed, logger *log.Logger) {
	if err := broker.Listen(func(msg *utils.BrokerMessage) bool {
		if ok, err := connections.WApply(msg.Receiver, func(_ string, conn *websocket.Conn) error {
			data, err := json.Marshal(msg)

			if err != nil {
				return err
			}

			if err := conn.WriteMessage(1, data); err != nil {
				return err
			}

			logger.Println("processmessages.success: sent message to connection")

			return nil
		}); !ok || err != nil {
			if !ok {
				logger.Println("processmessages.error: id does not exist")
			} else {
				logger.Println("processmessages.error: ", err)
			}

			return false
		}

		return true
	}, lock); err != nil {
		logger.Fatalln("processmessages.error: ", err)
	}
}
