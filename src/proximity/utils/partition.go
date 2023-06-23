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
	LatitudeMin    = -90
	LatitudeMax    = 90
	LongitudeMin   = -180
	LongitudeMax   = 180
)

// Calculate a partition string with a given depth from a given latitude and longitude
func calculatePartition(latitude float32, longitude float32, latitudeMin float32, latitudeMax float32, longitudeMin float32, longitudeMax float32, depth uint) (string, error) {
	// Base case
	if depth == 0 {
		return "", nil
	}

	if latitude < latitudeMin || latitude > latitudeMax || longitude < longitudeMin || longitude > longitudeMax {
		return "", errors.New("out of bounds")
	}

	id := 0

	// Recursive partitioning latitude
	midLat := (latitudeMin + latitudeMax) / 2

	var newLatMin float32
	var newLatMax float32

	if latitude < midLat {
		id += 0
		newLatMin = latitudeMin
		newLatMax = midLat
	} else {
		id += 1
		newLatMin = midLat
		newLatMax = latitudeMax
	}

	// Recursive partitioning longitude
	midLong := (longitudeMin + longitudeMax) / 2

	var newLongMin float32
	var newLongMax float32

	if longitude < midLong {
		id += 0
		newLongMin = longitudeMin
		newLongMax = midLong
	} else {
		id += 2
		newLongMin = midLong
		newLongMax = longitudeMax
	}

	suffix, err := calculatePartition(latitude, longitude, newLatMin, newLatMax, newLongMin, newLongMax, depth-1)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(id, suffix), nil
}

// Create a new partition from a latitude and longitude
func NewPartition(latitude float32, longitude float32) (*Partition, error) {
	encoded, err := calculatePartition(latitude, longitude, LatitudeMin, LatitudeMax, LongitudeMin, LongitudeMax, PartitionDepth)
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

// Find all surrounding partitions e.g. the surrounding 8 items
func (p *Partition) Surrounding() {
	// **** I do not know what I am doing here
}
