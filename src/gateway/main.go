package main

import (
	"log"
	"os"

	gwController "github.com/bengosborn/cue/gateway/src/controller"
	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	"github.com/joho/godotenv"
)

var addr = ":8080"
var workers = 10

func main() {
	logger := log.Logger{}

	if err := godotenv.Load(); err != nil {
		logger.Fatalln(err)
	}

	connections := gwUtils.NewConnections()
	queue := gwUtils.NewQueue(os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASSWORD"), os.Getenv("KAFKA_ENDPOINT"), os.Getenv("KAFKA_GROUP"), os.Getenv("KAFKA_TOPIC"), &logger)

	gwController.Start(addr, connections, workers, queue, &logger)
}
