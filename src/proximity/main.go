package main

import (
	"fmt"

	"github.com/bengosborn/cue/proximity/utils"
)

func main() {
	var longitude float32 = -60.0
	var latitude float32 = 20.0

	partition, err := utils.NewPartition(latitude, longitude)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(partition)
	}
}