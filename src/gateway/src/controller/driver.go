package gateway

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	utils "github.com/bengosborn/cue/utils"
)

// Start the server
func Start(addr string, connections *gwUtils.Connections, workers int, queue *utils.Queue, logger *log.Logger, process func(string, *gwUtils.Message) error) {
	messages := make(chan *utils.QueueMessage)

	http.HandleFunc("/", Handle(connections, logger, process))

	// Launch worker threads
	for i := 0; i < workers; i++ {
		go Worker(connections, messages, logger)
	}

	// Launch the event listener
	go Enqueue(queue, messages, logger)

	logger.Println("listening on address", addr)
	logger.Fatalln(http.ListenAndServe(addr, nil))
}
