package utils

import (
	"errors"
	"fmt"
	"strings"
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

	var newLongMin float32
	var newLongMax float32

	id := 0

	if longitude < midLong {
		id += 0
		newLongMin = longitudeMin
		newLongMax = midLong
	} else {
		id += 1
		newLongMin = midLong
		newLongMax = longitudeMax
	}

	// Recursive partitioning latitude
	midLat := (latitudeMin + latitudeMax) / 2

	var newLatMin float32
	var newLatMax float32

	if latitude < midLat {
		id += 0
		newLatMin = latitudeMin
		newLatMax = midLat
	} else {
		id += 2
		newLatMin = midLat
		newLatMax = latitudeMax
	}

	suffix, err := calculatePartition(latitude, longitude, newLongMin, newLongMax, newLatMin, newLatMax, depth-1)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(id, suffix), nil
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

// Check if one partition contains another
func (p *Partition) Contains(partition *Partition) bool {
	return strings.Contains(partition.encoded, p.encoded)
}
