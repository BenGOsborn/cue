package utils

import (
	"errors"
	"fmt"
)

type Partition struct {
	encoded string
}

// Depth of the partitioning e.g. the size of each partition
const (
	PartitionDepth = 10
	LongitudeMin   = -180
	LongitudeMax   = 180
	LatitudeMin    = -90
	LatitudeMax    = 90
)

// Calculate a partition string with a given depth from a given latitude and longitude
func calculatePartition(latitude float32, longitude float32, longitudeMin float32, longitudeMax float32, latitudeMin float32, latitudeMax float32, depth uint) (string, error) {
	// Base case
	if depth == 0 {
		return "", nil
	}

	if latitude < latitudeMin || latitude > latitudeMax || longitude < longitudeMin || longitude > longitudeMax {
		return "", errors.New("out of bounds")
	}

	// Recursive partitioning longitude
	midLong := (longitudeMin + longitudeMax) / 2

	var longDigit string
	var newLongMin float32
	var newLongMax float32

	if longitude < midLong {
		longDigit = "0"
		newLongMin = longitudeMin
		newLongMax = midLong
	} else {
		longDigit = "1"
		newLongMin = midLong
		newLongMax = longitudeMax
	}

	// Recursive partitioning latitude
	midLat := (latitudeMin + latitudeMax) / 2

	var latDigit string
	var newLatMin float32
	var newLatMax float32

	if latitude < midLat {
		latDigit = "0"
		newLatMin = latitudeMin
		newLatMax = midLat
	} else {
		latDigit = "1"
		newLatMin = midLat
		newLatMax = latitudeMax
	}

	suffix, err := calculatePartition(latitude, longitude, newLongMin, newLongMax, newLatMin, newLatMax, depth-1)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(longDigit, latDigit, suffix), nil
}

// Create a new partition from a latitude and longitude
func NewPartition(latitude float32, longitude float32) (*Partition, error) {
	encoded, err := calculatePartition(latitude, longitude, LongitudeMin, LongitudeMax, LatitudeMin, LatitudeMax, PartitionDepth)
	if err != nil {
		return nil, err
	}

	return &Partition{encoded: encoded}, nil
}

// Format the partition
func (p *Partition) String() string {
	return p.encoded
}
