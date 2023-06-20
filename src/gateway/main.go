package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	authController "github.com/bengosborn/cue/gateway/auth_controller"
	gwController "github.com/bengosborn/cue/gateway/gateway_controller"
	gwUtils "github.com/bengosborn/cue/gateway/utils"
	"github.com/bengosborn/cue/helpers"
	utils "github.com/bengosborn/cue/utils"
	"github.com/joho/godotenv"
)

var addr = "0.0.0.0:8080"

// Process a message
func Process(logger *log.Logger, queue *utils.Queue, session *gwUtils.Session, authenticator *gwUtils.Authenticator) func(string, *gwUtils.Message) error {
	return func(receiver string, msg *gwUtils.Message) error {
		logger.Println("process.received: received raw message")

		// Authenticate
		sessionData, err := session.Get(msg.SessionId)
		if err != nil {
			logger.Println(fmt.Sprint("process.error: ", err))

			return nil
		}

		user, err := authenticator.VerifyToken(sessionData.Token)
		if err != nil {
			logger.Println(fmt.Sprint("process.error: ", err))

			return nil
		}

		// Add to queue
		queueMsg := utils.QueueMessage{Receiver: receiver, User: user.Subject, EventType: msg.EventType, Body: msg.Body}
		queue.Send(&queueMsg)

		logger.Println("process.enqueued: added message to queue")

		return nil
	}
}

func main() {
	// Initialize environment
	logger := log.New(os.Stdout, "[Gateway] ", log.Ldate|log.Ltime)

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("../.env"); err != nil {
			logger.Fatalln(fmt.Scan("main.error", err))
		}
	}

	ctx := context.Background()
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Initialize data structures
	connections := gwUtils.NewConnections()
	defer connections.Close()

	redis, err := helpers.NewRedis(os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}
	defer redis.Close()

	queue, err := utils.NewQueue(ctx, os.Getenv("KAFKA_USERNAME"), os.Getenv("KAFKA_PASSWORD"), os.Getenv("KAFKA_ENDPOINT"), os.Getenv("KAFKA_TOPIC"), logger)
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}
	defer queue.Close()

	authenticator, err := gwUtils.NewAuthenticator(ctx, os.Getenv("AUTH0_DOMAIN"), os.Getenv("AUTH0_CALLBACK_URL"), os.Getenv("AUTH0_CLIENT_ID"), os.Getenv("AUTH0_CLIENT_SECRET"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}

	session := gwUtils.NewSession(ctx, redis)

	// Start server
	gwController.Attach(mux, "/ws", connections, queue, logger, Process(logger, queue, session, authenticator))
	authController.Attach(mux, "/auth", logger, session, authenticator)

	logger.Println("server listening on address", addr)
	logger.Fatalln(server.ListenAndServe())
}
