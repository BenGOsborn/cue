package gateway

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
)

// Start the server
func Start(addr string, connections *gwUtils.Connections, workers int, queue *gwUtils.Queue, logger *log.Logger) {
	messages := make(chan *gwUtils.QueueMessage)

	http.HandleFunc("/", Handle(connections, logger, Process(logger)))

	// Launch worker threads
	for i := 0; i < workers; i++ {
		go Worker(connections, messages, logger)
	}

	// Launch the event listener
	go Enqueue(queue, messages, logger)

	logger.Println("listening on address", addr)
	logger.Fatalln(http.ListenAndServe(addr, nil))
}
