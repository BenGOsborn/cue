package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bengosborn/cue/helpers"
	pUtils "github.com/bengosborn/cue/proximity/utils"
	"github.com/bengosborn/cue/utils"
	"github.com/joho/godotenv"
)

var timeout = 5 * time.Minute

func main() {
	logger := log.New(os.Stdout, "[Gateway] ", log.Ldate|log.Ltime)

	// Initialize environment
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("../.env"); err != nil {
			logger.Fatalln(err)
		}
	}

	ctx := context.Background()

	redis, err := helpers.NewRedis(os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Fatalln(err)
	}
	defer redis.Close()

	// Create location
	var lat float32 = 20.0
	var long float32 = -60.0
	user1 := "test123"
	user2 := "test456"
	locationId := "1"

	scanner := bufio.NewScanner(os.Stdin)

	lock, err := utils.NewResourceLockDistributed(ctx, redis, timeout)
	if err != nil {
		logger.Fatalln(err)
	}

	location1 := pUtils.NewLocation(ctx, redis, lock, locationId)
	if err := location1.Upsert(user1, lat, long); err != nil {
		logger.Fatalln(err)
	}

	location2 := pUtils.NewLocation(ctx, redis, lock, locationId)
	if err := location2.Upsert(user2, lat, long); err != nil {
		logger.Fatalln(err)
	}

	scanner.Scan()

	if err := location1.Sync(); err != nil {
		logger.Fatalln(err)
	}

	scanner.Scan()

	if err := location2.Sync(); err != nil {
		logger.Fatalln(err)
	}

	out, err := location1.Nearby(user1, 1)
	if err != nil {
		logger.Fatalln(err)
	} else {
		fmt.Println(out)
	}

	out, err = location2.Nearby(user1, 1)
	if err != nil {
		logger.Fatalln(err)
	} else {
		fmt.Println(out)
	}

	scanner.Scan()

	for e := location2.EventStack.Front(); e != nil; e = e.Next() {
		fmt.Println("Location 2 event", e.Value)
	}

	location2.Remove(user1)
	location2.Remove(user2)
	if err := location2.Sync(); err != nil {
		logger.Fatalln(err)
	}

	out, err = location2.Nearby(user1, 1)
	if err != nil {
		logger.Fatalln(err)
	} else {
		fmt.Println(out)
	}
}
