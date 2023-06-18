package gateway

import (
	"log"

	gateway "github.com/bengosborn/cue/gateway/src/utils"
)

// Process a message
func Process(logger *log.Logger) func(*gateway.Message) error {
	return func(m *gateway.Message) error {
		logger.Println(m)

		return nil
	}
}
