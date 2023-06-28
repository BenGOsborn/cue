package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bengosborn/cue/helpers"
	"github.com/bengosborn/cue/proximity/controller"
	pUtils "github.com/bengosborn/cue/proximity/utils"
	"github.com/bengosborn/cue/utils"
	"github.com/joho/godotenv"
)

const (
	lockTimeout     = 5 * time.Minute
	locationTimeout = 5 * time.Minute
	serviceId       = "proximity:main"
)

func main() {
	logger := log.New(os.Stdout, "[Gateway] ", log.Ldate|log.Ltime)
	ctx := context.Background()

	// Initialize environment
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("../.env"); err != nil {
			logger.Fatalln(err)
		}
	}

	redis, err := helpers.NewRedis(os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Fatalln(err)
	}
	defer redis.Close()

	lock, err := utils.NewResourceLockDistributed(ctx, redis, lockTimeout)
	if err != nil {
		logger.Fatalln(err)
	}

	brokerIn := utils.NewBrokerRedis(ctx, redis, os.Getenv("REDIS_PROXIMITY_CHANNEL_IN"), serviceId)
	brokerOut := utils.NewBrokerRedis(ctx, redis, os.Getenv("REDIS_GATEWAY_CHANNEL_IN"), serviceId)

	location := pUtils.NewLocation(ctx, serviceId, locationTimeout, redis, lock)

	logger.Println("starting proximity service...")
	controller.Controller(ctx, location, brokerIn, brokerOut, lock, logger)
}
