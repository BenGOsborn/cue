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
	// PartitionDepth = 10
	PartitionDepth = 3
	LatMin         = -90
	LatMax         = 90
	LongMin        = -180
	LongMax        = 180
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
func partition(lat float32, long float32, latMin float32, latMax float32, longMin float32, longMax float32, depth uint) (string, *[]*Chunk, error) {
	buffer := strings.Builder{}
	chunks := make([]*Chunk, depth)

	var recurse func(float32, float32, float32, float32, uint) error
	recurse = func(latMin float32, latMax float32, longMin float32, longMax float32, depth uint) error {
		// Base case
		if depth <= 0 {
			return nil
		}

		if lat < latMin || lat > latMax || long < longMin || long > longMax {
			return errors.New("out of bounds")
		}

		// Recursive partitioning latitude
		midLat := (latMin + latMax) / 2

		y := 0
		var newLatMin float32
		var newLatMax float32

		if lat < midLat {
			newLatMin = latMin
			newLatMax = midLat
		} else {
			y = 1
			newLatMin = midLat
			newLatMax = latMax
		}

		// Recursive partitioning longitude
		midLong := (longMin + longMax) / 2

		x := 0
		var newLongMin float32
		var newLongMax float32

		if long < midLong {
			newLongMin = longMin
			newLongMax = midLong
		} else {
			x = 1
			newLongMin = midLong
			newLongMax = longMax
		}

		// Write the data
		depth -= 1
		chunks[depth] = &Chunk{y: y, x: x}

		if _, err := buffer.WriteString(fmt.Sprint(2*y + x)); err != nil {
			return err
		}

		return recurse(newLatMin, newLatMax, newLongMin, newLongMax, depth)
	}

	if err := recurse(latMin, latMax, longMin, longMax, depth); err != nil {
		return "", nil, err
	}

	return buffer.String(), &chunks, nil
}

// Create a new partition from a latitude and longitude
func NewPartitionFromCoords(lat float32, long float32) (*Partition, error) {
	encoded, chunks, err := partition(lat, long, LatMin, LatMax, LongMin, LongMax, PartitionDepth)
	if err != nil {
		return nil, err
	}

	return &Partition{encoded: encoded, chunks: chunks}, nil
}

// Create a new partition from chunks
func NewPartitonFromChunks(chunks *[]*Chunk) (*Partition, error) {
	buffer := strings.Builder{}

	for i := len(*chunks) - 1; i >= 0; i-- {
		chunk := (*chunks)[i]

		if _, err := buffer.WriteString(fmt.Sprint(2*chunk.y + chunk.x)); err != nil {
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

	for i, chunk := range *chunks {
		newY := chunk.y
		if remainderY != 0 {
			temp := newY + remainderY

			if newY == 0 {
				newY = 1
			} else {
				newY = 0
			}

			if temp == 0 || temp == 1 {
				remainderY = 0
			}
		}

		newX := chunk.x
		if remainderX != 0 {
			temp := newX + remainderX

			if newX == 0 {
				newX = 1
			} else {
				newX = 0
			}

			if temp == 0 || temp == 1 {
				remainderX = 0
			}
		}

		newChunks[i] = &Chunk{x: newX, y: newY}
	}

	newPartition, err := NewPartitonFromChunks(&newChunks)
	if err != nil {
		return nil, err
	}

	return newPartition, nil
}

type QueueNode struct {
	remaining uint
	partition *Partition
}

// Add a new partition
func addPartition(node *QueueNode, queue *list.List, seen *map[string]bool, out *[]*Partition, direction Direction) error {
	partition, err := translate(node.partition, direction)
	if err != nil {
		return err
	}

	// No duplicates
	if _, ok := (*seen)[partition.encoded]; ok {
		return nil
	}
	(*seen)[partition.encoded] = true

	*out = append(*out, partition)

	// Add to queue
	remaining := node.remaining - 1

	if remaining > 0 {
		queue.PushBack(&QueueNode{remaining: remaining, partition: partition})
	}

	return nil
}

// Find all surrounding partitions within a given radius
func (p *Partition) Surrounding(radius uint) (*[]*Partition, error) {
	seen := make(map[string]bool)
	queue := list.New()
	queue.PushBack(&QueueNode{remaining: radius, partition: p})

	out := make([]*Partition, 0)

	// BFS for surrounding partitons
	for queue.Len() > 0 {
		current := queue.Front()
		node := current.Value.(*QueueNode)
		queue.Remove(current)

		// Create new surrounding partitions
		if err := addPartition(node, queue, &seen, &out, DirUp); err != nil {
			return nil, err
		}
		if err := addPartition(node, queue, &seen, &out, DirDown); err != nil {
			return nil, err
		}
		if err := addPartition(node, queue, &seen, &out, DirLeft); err != nil {
			return nil, err
		}
		if err := addPartition(node, queue, &seen, &out, DirRight); err != nil {
			return nil, err
		}
	}

	return &out, nil
}
