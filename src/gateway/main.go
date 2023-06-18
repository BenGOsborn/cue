package main

import (
	"log"
	"os"

	gwController "github.com/bengosborn/cue/gateway/src/controller"
	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	utils "github.com/bengosborn/cue/utils"
	"github.com/joho/godotenv"
)

var addr = ":8080"
var workers = 10

func main() {
	logger := log.New(os.Stdout, "[Gateway] ", log.Ldate|log.Ltime)

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("../.env"); err != nil {
			logger.Fatalln(err)
		}
	}

	queue := utils.NewQueue(os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASSWORD"), os.Getenv("KAFKA_ENDPOINT"), os.Getenv("KAFKA_TOPIC"), logger)
	defer queue.Close()

	connections := gwUtils.NewConnections()
	defer connections.Close()

	gwController.Start(addr, connections, workers, queue, logger, gwController.Process(logger, queue))
}
