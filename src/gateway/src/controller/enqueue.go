package gateway

import (
	"log"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
)

// Enqueue an element from an event listener
func Enqueue(queue *gwUtils.Queue, messages chan<- *gwUtils.QueueMessage, logger *log.Logger) {
	if err := queue.Listen(func(qm *gwUtils.QueueMessage) error {
		messages <- qm

		return nil
	}); err != nil {
		logger.Fatalln(err)
	}
}
