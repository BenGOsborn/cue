package controller

import (
	"fmt"
	"log"

	"github.com/bengosborn/cue/utils"
)

// Routing logic for all broker messages
func Process(broker utils.Broker, lock *utils.ResourceLockDistributed, logger *log.Logger) {
	if err := broker.Listen(func(msg *utils.BrokerMessage) bool {
		switch msg.EventType {
		case (utils.ProximitySendLocation):
			fallthrough
		case (utils.ProximityRequestNearby):
			fallthrough
		default:
			return false
		}

		return true

	}, lock); err != nil {
		logger.Fatalln(fmt.Sprint("process.error: ", err))
	}
}
