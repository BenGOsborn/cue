package utils

import (
	"container/list"
	"errors"
	"fmt"
	"strings"
)

type Chunk struct {
	y int
	x int
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

// Translate directions
type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
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

		down := 0
		var newLatMin float32
		var newLatMax float32

		if latitude < midLat {
			newLatMin = latitudeMin
			newLatMax = midLat
		} else {
			down = 1
			newLatMin = midLat
			newLatMax = latitudeMax
		}

		// Recursive partitioning longitude
		midLong := (longitudeMin + longitudeMax) / 2

		right := 0
		var newLongMin float32
		var newLongMax float32

		if longitude < midLong {
			newLongMin = longitudeMin
			newLongMax = midLong
		} else {
			right = 1
			newLongMin = midLong
			newLongMax = longitudeMax
		}

		// Write the data
		depth -= 1
		chunks[depth] = &Chunk{y: down, x: right}

		if _, err := buffer.WriteString(fmt.Sprint(down)); err != nil {
			return err
		}

		if _, err := buffer.WriteString(fmt.Sprint(right)); err != nil {
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
func NewPartitionFromCoords(latitude float32, longitude float32) (*Partition, error) {
	encoded, chunks, err := partition(latitude, longitude, LatitudeMin, LatitudeMax, LongitudeMin, LongitudeMax, PartitionDepth)
	if err != nil {
		return nil, err
	}

	return &Partition{encoded: encoded, chunks: chunks}, nil
}

// Create a new partition from chunks
func NewPartitonFromChunks(chunks *[]*Chunk) (*Partition, error) {
	buffer := strings.Builder{}

	for _, chunk := range *chunks {
		if _, err := buffer.WriteString(fmt.Sprint(chunk.y)); err != nil {
			return nil, err
		}

		if _, err := buffer.WriteString(fmt.Sprint(chunk.x)); err != nil {
			return nil, err
		}

		if _, err := buffer.WriteString(" "); err != nil {
			return nil, err
		}
	}

	return &Partition{chunks: chunks, encoded: buffer.String()}, nil
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
func translate(partition *Partition, direction Direction) (*Partition, error) {
	var remainderY int
	var remainderX int

	switch direction {
	case DirUp:
		remainderY = -1
		remainderX = 0
	case DirDown:
		remainderY = 1
		remainderX = 0
	case DirRight:
		remainderY = 0
		remainderX = 1
	case DirLeft:
		remainderY = 0
		remainderX = -1
	default:
		return nil, errors.New("invalid direction")
	}

	chunks := partition.chunks
	newChunks := make([]*Chunk, len(*chunks))

	for i := len(*chunks) - 1; i >= 0; i-- {
		chunk := (*chunks)[i]

		newY := chunk.y + remainderY
		if newY > 1 {
			newY = 0
		} else if newY < 0 {
			newY = 1
		} else {
			remainderY = 0
		}

		newX := chunk.x + remainderX
		if newX > 1 {
			newX = 0
		} else if newX < 0 {
			newX = 1
		} else {
			remainderX = 0
		}

		newChunks[i] = &Chunk{x: newX, y: newY}
	}

	newPartition, err := NewPartitonFromChunks(&newChunks)
	if err != nil {
		return nil, err
	}

	return newPartition, nil
}

type Node struct {
	remaining uint
	partition *Partition
}

// Find all surrounding partitions within a given radius
func (p *Partition) Surrounding(radius uint) error {
	seen := make(map[string]bool)
	queue := list.New()
	queue.PushBack(&Node{remaining: radius, partition: p})

	// BFS for surrounding partitons
	for queue.Len() > 0 {
		current := queue.Front()
		node := current.Value.(*Node)
		queue.Remove(current)

		if _, ok := seen[node.partition.encoded]; ok || node.remaining == 0 {
			continue
		}
		seen[node.partition.encoded] = true

		// Create new partitions
		{
			partition, err := translate(node.partition, DirUp)
			if err != nil {
				return err
			}
			queue.PushBack(&Node{remaining: node.remaining - 1, partition: partition})
		}
		{
			partition, err := translate(node.partition, DirDown)
			if err != nil {
				return err
			}
			queue.PushBack(&Node{remaining: node.remaining - 1, partition: partition})
		}
		{
			partition, err := translate(node.partition, DirLeft)
			if err != nil {
				return err
			}
			queue.PushBack(&Node{remaining: node.remaining - 1, partition: partition})
		}
		{
			partition, err := translate(node.partition, DirRight)
			if err != nil {
				return err
			}
			queue.PushBack(&Node{remaining: node.remaining - 1, partition: partition})
		}
	}

	return nil
}
