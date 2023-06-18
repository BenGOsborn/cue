package gateway

import (
	"fmt"
	"log"

	utils "github.com/bengosborn/cue/utils"
)

// Enqueue an element from an event listener
func Enqueue(queue *utils.Queue, messages chan<- *utils.QueueMessage, logger *log.Logger) {
	if err := queue.Listen(func(qm *utils.QueueMessage) error {
		messages <- qm

		logger.Println("Enqueue.enqueued: added message from queue")

		return nil
	}); err != nil {
		logger.Fatalln(fmt.Sprint("Enqueue.error: ", err))
	}
}
