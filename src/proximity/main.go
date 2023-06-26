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
	user2 := "test456"

	lock1, err := utils.NewResourceLockDistributed(ctx, redis, timeout)
	if err != nil {
		fmt.Println(err)
		return
	}
	location1 := pUtils.NewLocation(ctx, redis, lock1)

	if err := location1.Upsert(user1, lat, long); err != nil {
		fmt.Println(err)
		return
	}

	lock2, err := utils.NewResourceLockDistributed(ctx, redis, timeout)
	if err != nil {
		fmt.Println(err)
		return
	}
	location2 := pUtils.NewLocation(ctx, redis, lock2)

	if err := location2.Upsert(user2, lat, long); err != nil {
		fmt.Println(err)
		return
	}

	location2.Remove(user2)

	location2.Merge(location1)

	out, err := location1.Nearby(user1, 1)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(out)
	}

	out, err = location2.Nearby(user1, 1)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(out)
	}

}
