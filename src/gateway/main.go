package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	authController "github.com/bengosborn/cue/gateway/src/auth_controller"
	gwController "github.com/bengosborn/cue/gateway/src/gateway_controller"
	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	utils "github.com/bengosborn/cue/utils"
	"github.com/joho/godotenv"
)

var addr = ":8080"

// Process a message
func Process(logger *log.Logger, queue *utils.Queue, authenticator *utils.Authenticator, client *utils.Redis) func(string, *gwUtils.Message) error {
	return func(id string, msg *gwUtils.Message) error {
		logger.Println("Process.received: received raw message")

		// Authenticate
		// msg.Auth -> authenticate this once -> cache it -> pull from cache

		// Add to queue
		queueMsg := utils.QueueMessage{Receiver: id, Type: msg.Type, Body: msg.Body}
		queue.Send(&queueMsg)

		logger.Println("Process.enqueued: added message to queue")

		return nil
	}
}

func main() {
	logger := log.New(os.Stdout, "[Gateway] ", log.Ldate|log.Ltime)
	ctx := context.Background()
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("../.env"); err != nil {
			logger.Fatalln(fmt.Scan("main.error", err))
		}
	}

	connections := gwUtils.NewConnections()
	defer connections.Close()

	client, err := utils.NewRedis(ctx, os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}
	defer client.Close()

	queue, err := utils.NewQueue(ctx, os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASSWORD"), os.Getenv("KAFKA_ENDPOINT"), os.Getenv("KAFKA_TOPIC"), logger)
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}
	defer queue.Close()

	authenticator, err := utils.NewAuthenticator(ctx, os.Getenv("AUTH0_DOMAIN"), os.Getenv("AUTH0_CALLBACK_URL"), os.Getenv("AUTH0_CLIENT_ID"), os.Getenv("AUTH0_CLIENT_SECRET"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}

	gwController.Attach(mux, "/ws", connections, queue, logger, Process(logger, queue, authenticator, client))
	authController.Attach(mux, "/auth", "/auth/callback", logger, client, authenticator)

	fmt.Println("server listening on address", addr)
	logger.Fatalln(server.ListenAndServe())
}
