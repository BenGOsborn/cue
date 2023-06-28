package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	authController "github.com/bengosborn/cue/gateway/auth_controller"
	gwController "github.com/bengosborn/cue/gateway/gateway_controller"
	gwUtils "github.com/bengosborn/cue/gateway/utils"
	"github.com/bengosborn/cue/helpers"
	utils "github.com/bengosborn/cue/utils"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

const (
	addr        = "0.0.0.0:8080"
	lockTimeout = 5 * time.Minute
)

// Process a message
func Process(logger *log.Logger, brokerProximity utils.Broker, session *gwUtils.Session, authenticator *gwUtils.Authenticator) func(string, *gwUtils.Message) error {
	return func(receiver string, msg *gwUtils.Message) error {
		logger.Println("process.received: received raw message")

		// Authenticate
		sessionData, err := session.Get(msg.SessionId)
		if err != nil {
			logger.Println("process.error: ", err)

			return err
		}

		user, err := authenticator.VerifyToken(sessionData.Token)
		if err != nil {
			logger.Println("process.error: ", err)

			return err
		}

		// Send to broker
		brokerMsgId := uuid.NewString()

		switch msg.EventType {
		case utils.ProximityRequestNearby, utils.ProximitySendLocation:
			brokerProximity.Send(&utils.BrokerMessage{Id: brokerMsgId, Receiver: receiver, User: user.Subject, EventType: msg.EventType, Body: msg.Body})

			logger.Println("process.sent: sent message to proximity broker")
		default:
			logger.Println("process.error: invalid event type")
		}

		return nil
	}
}

func main() {
	// Initialize environment
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

	// Initialize data structures
	connections := gwUtils.NewConnections()
	defer connections.Close()

	redis, err := helpers.NewRedis(os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}
	defer redis.Close()

	brokerIn := utils.NewBrokerRedis(ctx, redis, os.Getenv("REDIS_GATEWAY_CHANNEL_IN"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}

	brokerProximity := utils.NewBrokerRedis(ctx, redis, os.Getenv("REDIS_PROXIMITY_CHANNEL_IN"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}

	authenticator, err := gwUtils.NewAuthenticator(ctx, os.Getenv("AUTH0_DOMAIN"), os.Getenv("AUTH0_CALLBACK_URL"), os.Getenv("AUTH0_CLIENT_ID"), os.Getenv("AUTH0_CLIENT_SECRET"))
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}

	lock, err := utils.NewResourceLockDistributed(ctx, redis, lockTimeout)
	if err != nil {
		logger.Fatalln(fmt.Scan("main.error", err))
	}

	session := gwUtils.NewSession(ctx, redis)

	// Start server
	gwController.Attach(mux, "/ws", connections, brokerIn, lock, logger, Process(logger, brokerProximity, session, authenticator))
	authController.Attach(mux, "/auth", logger, session, authenticator)

	logger.Println("server listening on address", addr)
	logger.Fatalln(server.ListenAndServe())
}
