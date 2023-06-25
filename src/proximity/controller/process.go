package controller

import (
	"fmt"
	"log"

	"github.com/bengosborn/cue/utils"
)

// Routing logic for all broker messages
func ProcessMessages(broker utils.Broker, lock *utils.ResourceLockDistributed, logger *log.Logger) {
	if err := broker.Listen(func(msg *utils.BrokerMessage) bool {
		switch msg.EventType {
		case (utils.ProximitySendLocation):
			fallthrough
		case (utils.ProximityRequestNearby):
			fallthrough
		default:
			return false
		}

		// **** Might need to use batching depending on the size of the messages to come up with some state locally and then write it to Redis every few seconds
		// for consistency (read from Redis, write locally until timeout where we aggregate the messages and send them up)

		return true

	}, lock); err != nil {
		logger.Fatalln(fmt.Sprint("processmessages.error: ", err))
	}
}
