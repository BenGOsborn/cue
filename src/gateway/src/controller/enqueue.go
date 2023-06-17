package gateway

import (
	"log"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
)

// Enqueue an element from an event listener
// **** I need to add support for JSON here across all of the items
func Enqueue(queue *gwUtils.Queue, messages chan<- gwUtils.Message, logger *log.Logger) {
	if err := queue.Listen(func(s string) error {
		logger.Println("here is my new item:", s)

		return nil
	}); err != nil {
		logger.Fatalln(err)
	}
}
