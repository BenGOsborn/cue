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
	logger := log.New(os.Stdout, "[Gateway] ", log.Ldate|log.Ltime)

	if err := godotenv.Load(); err != nil {
		logger.Fatalln(err)
	}

	queue := gwUtils.NewQueue(os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASSWORD"), os.Getenv("KAFKA_ENDPOINT"), os.Getenv("KAFKA_TOPIC"), logger)
	connections := gwUtils.NewConnections()

	gwController.Start(addr, connections, workers, queue, logger)
}
