package gateway

import (
	"log"

	gateway "github.com/bengosborn/cue/gateway/src/utils"
	utils "github.com/bengosborn/cue/utils"
)

// Process a message
func Process(logger *log.Logger, queue *utils.Queue) func(string, *gateway.Message) error {
	return func(id string, msg *gateway.Message) error {
		// Authenticate
		// msg.Auth

		// Add to queue
		queueMsg := utils.QueueMessage{Receiver: id, Type: msg.Type, Body: msg.Body}
		queue.Send(&queueMsg)

		return nil
	}
}
