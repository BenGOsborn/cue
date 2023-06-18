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
func Process(logger *log.Logger, queue *utils.Queue, authenticator *utils.Authenticator) func(string, *gwUtils.Message) error {
	return func(receiver string, msg *gwUtils.Message) error {
		logger.Println("process.received: received raw message")

		// Authenticate
		user, err := authenticator.VerifyIDToken(msg.Auth)

		if err != nil {
			return nil
		}

		// Add to queue
		queueMsg := utils.QueueMessage{Receiver: receiver, User: user, EventType: msg.EventType, Body: msg.Body}
		queue.Send(&queueMsg)

		logger.Println("process.enqueued: added message to queue")

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

	redis, err := utils.NewRedis(ctx, os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}
	defer redis.Close()

	queue, err := utils.NewQueue(ctx, os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASSWORD"), os.Getenv("KAFKA_ENDPOINT"), os.Getenv("KAFKA_TOPIC"), logger)
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}
	defer queue.Close()

	authenticator, err := utils.NewAuthenticator(ctx, os.Getenv("AUTH0_DOMAIN"), os.Getenv("AUTH0_CALLBACK_URL"), os.Getenv("AUTH0_CLIENT_ID"), os.Getenv("AUTH0_CLIENT_SECRET"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}

	gwController.Attach(mux, "/ws", connections, []*utils.Queue{queue}, logger, Process(logger, queue, authenticator))
	authController.Attach(mux, "/auth", "/auth/callback", logger, redis, authenticator)

	fmt.Println("server listening on address", addr)
	logger.Fatalln(server.ListenAndServe())
}
