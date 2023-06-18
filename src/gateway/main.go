package main

import (
	"fmt"
	"log"
	"os"

	gwController "github.com/bengosborn/cue/gateway/src/controller"
	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	utils "github.com/bengosborn/cue/utils"
	"github.com/joho/godotenv"
)

var addr = ":8080"
var workers = 10

// Process a message
func Process(logger *log.Logger, queue *utils.Queue) func(string, *gwUtils.Message) error {
	return func(id string, msg *gwUtils.Message) error {
		logger.Println("Process.received: received raw message")

		// Authenticate
		// msg.Auth

		// Add to queue
		queueMsg := utils.QueueMessage{Receiver: id, Type: msg.Type, Body: msg.Body}
		queue.Send(&queueMsg)

		logger.Println("Process.enqueued: added message to queue")

		return nil
	}
}

func main() {
	logger := log.New(os.Stdout, "[Gateway] ", log.Ldate|log.Ltime)

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("../.env"); err != nil {
			logger.Fatalln(fmt.Scan("main.error", err))
		}
	}

	queue, err := utils.NewQueue(os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASSWORD"), os.Getenv("KAFKA_ENDPOINT"), os.Getenv("KAFKA_TOPIC"), logger)
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}
	defer queue.Close()

	connections := gwUtils.NewConnections()
	defer connections.Close()

	gwController.Start(addr, connections, workers, queue, logger, Process(logger, queue))
}
