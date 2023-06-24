package utils

import (
	"errors"
	"fmt"
	"strings"
)

type Chunk struct {
	right bool
	down  bool
}

type Partition struct {
	encoded string
	chunks  *[]*Chunk
}

// Depth of the partitioning e.g. the size of each partition
const (
	PartitionDepth = 10
	LatitudeMin    = -90
	LatitudeMax    = 90
	LongitudeMin   = -180
	LongitudeMax   = 180
)

// Partition a given latitude and longitude
func partition(latitude float32, longitude float32, latitudeMin float32, latitudeMax float32, longitudeMin float32, longitudeMax float32, depth uint) (string, *[]*Chunk, error) {
	buffer := strings.Builder{}
	chunks := make([]*Chunk, depth)

	var recurse func(float32, float32, float32, float32, uint) error
	recurse = func(latitudeMin float32, latitudeMax float32, longitudeMin float32, longitudeMax float32, depth uint) error {
		// Base case
		if depth <= 0 {
			return nil
		}

		if latitude < latitudeMin || latitude > latitudeMax || longitude < longitudeMin || longitude > longitudeMax {
			return errors.New("out of bounds")
		}

		// Recursive partitioning latitude
		midLat := (latitudeMin + latitudeMax) / 2

		down := false
		var newLatMin float32
		var newLatMax float32

		if latitude < midLat {
			newLatMin = latitudeMin
			newLatMax = midLat
		} else {
			down = true
			newLatMin = midLat
			newLatMax = latitudeMax
		}

		// Recursive partitioning longitude
		midLong := (longitudeMin + longitudeMax) / 2

		right := false
		var newLongMin float32
		var newLongMax float32

		if longitude < midLong {
			newLongMin = longitudeMin
			newLongMax = midLong
		} else {
			right = true
			newLongMin = midLong
			newLongMax = longitudeMax
		}

		// Write the data
		depth -= 1
		chunks[depth] = &Chunk{down: down, right: right}

		downDigit := "0"
		if down {
			downDigit = "1"
		}
		if _, err := buffer.WriteString(downDigit); err != nil {
			return err
		}

		rightDigit := "0"
		if right {
			rightDigit = "1"
		}
		if _, err := buffer.WriteString(rightDigit); err != nil {
			return err
		}

		if _, err := buffer.WriteString(" "); err != nil {
			return err
		}

		return recurse(newLatMin, newLatMax, newLongMin, newLongMax, depth)
	}

	if err := recurse(latitudeMin, latitudeMax, longitudeMin, longitudeMax, depth); err != nil {
		return "", nil, err
	}

	return buffer.String(), &chunks, nil
}

// Create a new partition from a latitude and longitude
func NewPartition(latitude float32, longitude float32) (*Partition, error) {
	encoded, chunks, err := partition(latitude, longitude, LatitudeMin, LatitudeMax, LongitudeMin, LongitudeMax, PartitionDepth)
	if err != nil {
		return nil, err
	}

	fmt.Println(encoded)
	fmt.Println(chunks)

	return &Partition{encoded: encoded, chunks: chunks}, nil
}

// Format the partition
func (p *Partition) String() string {
	return p.encoded
}

// Check if one partition contains another
func (p *Partition) Contains(partition *Partition) bool {
	return strings.Contains(partition.encoded, p.encoded)
}

// Translate a partition string in some direction
func translate() {

}

// Find all surrounding partitions e.g. the surrounding 8 items
func (p *Partition) Surrounding() {
	// **** I do not know what I am doing here

	// **** Ground rules - if we need to go back from 0 we need to go to the previous, we rotate back to the end of the string in our search

	// **** We actually need a trie structure for this instead
}
