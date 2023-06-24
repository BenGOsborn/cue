package main

import (
	"fmt"

	"github.com/bengosborn/cue/proximity/utils"
)

func main() {
	var longitude float32 = -60.0
	var latitude float32 = 20.0

	partition, err := utils.NewPartitionFromCoords(latitude, longitude)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(partition)
	}

	out, err := partition.Surrounding(1)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(out, len(*out))
	}

	// utils.Test(partition)

}
