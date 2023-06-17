package gateway

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
)

// **** So now we need to create a facade for a message queue (which supports fan out) and which sends all of its messages into one of these functions to be sorted
// into another channel for where the message should go and what category it should be assigned ??? (this needs to be scalable)
// Raw messagw from service -> message queue processing channel -> processed ready to be sent to message channel / broadcast to user channel / broadcast to all chanel

// **** For efficiency, we will have in "express" Kafka channel which when the message is sent this server id is attached to it, and then when the message is sent back
// it will go directly to a queue only listened to by this server which will process it - if the message does not belong to this user, we will add it to a global queue
// which is listened to by all servers.
// Also remove the broadcast to all - it will NOT work in a distributed environment and is not needed.

// Start the server
func Start(addr string, connections *gwUtils.Connections, workers int, queue *gwUtils.Queue, logger *log.Logger) {
	messages := make(chan gwUtils.Message)

	http.HandleFunc("/", Handle(connections, logger))

	// Launch worker threads
	for i := 0; i < workers; i++ {
		go Worker(connections, messages, logger)
	}

	// Launch the event listener
	go Enqueue(queue, messages, logger)

	logger.Println("listening on address", addr)
	logger.Fatal(http.ListenAndServe(addr, nil))
}
