package main

import (
	"fmt"

	"github.com/bengosborn/cue/proximity/utils"
)

func main() {
	var lat float32 = 20.0
	var long float32 = -60.0
	user1 := "test123"
	user2 := "test456"

	location := utils.NewLocation()

	if err := location.Upsert(user1, lat, long); err != nil {
		fmt.Println(err)
		return
	}

	if err := location.Upsert(user2, lat, long); err != nil {
		fmt.Println(err)
		return
	}

	out, err := location.Nearby(user1, 0)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(out)
	}

	partition, err := utils.NewPartitionFromEncoded("21212")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(partition)
	}
}
