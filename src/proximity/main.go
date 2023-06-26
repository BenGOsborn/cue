package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bengosborn/cue/helpers"
	pUtils "github.com/bengosborn/cue/proximity/utils"
	"github.com/bengosborn/cue/utils"
	"github.com/joho/godotenv"
)

var timeout = 5 * time.Minute

func main() {
	// Initialize environment
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("../.env"); err != nil {
			fmt.Println(err)
			return
		}
	}

	ctx := context.Background()

	redis, err := helpers.NewRedis(os.Getenv("REDIS_URL"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer redis.Close()

	// Create location
	var lat float32 = 20.0
	var long float32 = -60.0
	user1 := "test123"

	lock, err := utils.NewResourceLockDistributed(ctx, redis, timeout)
	if err != nil {
		fmt.Println(err)
		return
	}

	location := pUtils.NewLocation(ctx, redis, lock, "1")

	if err := location.Upsert(user1, lat, long); err != nil {
		fmt.Println(err)
		return
	}

	if err := location.Sync(); err != nil {
		fmt.Println(err)
		return
	}

	out, err := location.Nearby(user1, 1)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(out)
	}
}
