package gateway

import (
	"log"

	gateway "github.com/bengosborn/cue/gateway/src/utils"
)

// Process a message
func Process(message gateway.Message, logger *log.Logger) {
	log.Println(message.Id + ": " + message.Message)
}
