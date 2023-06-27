package utils

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type chunk struct {
	Y int `json:"y"`
	X int `json:"x"`
}

type Partition struct {
	Encoded string
	Chunks  *[]*chunk
}

// Depth of the partitioning e.g. the size of each partition
const (
	PartitionDepth = 10
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
func partition(lat float32, long float32, latMin float32, latMax float32, longMin float32, longMax float32, depth uint) (string, *[]*chunk, error) {
	buffer := strings.Builder{}
	chunks := make([]*chunk, depth)

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
		chunks[depth] = &chunk{Y: y, X: x}

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

	return &Partition{Encoded: encoded, Chunks: chunks}, nil
}

// Create a new partition from chunks
func NewPartitonFromChunks(chunks *[]*chunk) (*Partition, error) {
	buffer := strings.Builder{}

	for i := len(*chunks) - 1; i >= 0; i-- {
		chunk := (*chunks)[i]

		if _, err := buffer.WriteString(fmt.Sprint(2*chunk.Y + chunk.X)); err != nil {
			return nil, err
		}
	}

	return &Partition{Chunks: chunks, Encoded: buffer.String()}, nil
}

// Create a new partition from a encoded string
func NewPartitionFromEncoded(encoded string) (*Partition, error) {
	chunks := make([]*chunk, len(encoded))

	for i, char := range encoded {
		var chn chunk

		switch string(char) {
		case "0":
			chn = chunk{X: 0, Y: 0}
		case "1":
			chn = chunk{X: 1, Y: 0}
		case "2":
			chn = chunk{X: 0, Y: 1}
		case "3":
			chn = chunk{X: 1, Y: 1}
		default:
			return nil, errors.New("invalid string character")
		}

		chunks[len(encoded)-i-1] = &chn
	}

	return &Partition{Encoded: encoded, Chunks: &chunks}, nil
}

// Format the partition
func (p *Partition) String() string {
	return p.Encoded
}

// Check if one partition contains another
func (p *Partition) Contains(partition *Partition) bool {
	return strings.Contains(partition.Encoded, p.Encoded)
}

// Translate a partition string in some direction
func (p *Partition) Translate(direction Direction) (*Partition, error) {
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
		panic("invalid direction")
	}

	chunks := p.Chunks
	newChunks := make([]*chunk, len(*chunks))

	for i, chn := range *chunks {
		newY := chn.Y
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

		newX := chn.X
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

		newChunks[i] = &chunk{X: newX, Y: newY}
	}

	newPartition, err := NewPartitonFromChunks(&newChunks)
	if err != nil {
		return nil, err
	}

	return newPartition, nil
}

type queueNode struct {
	remaining int
	partition *Partition
}

// Find all nearby partitions within a given radius
func (p *Partition) Nearby(radius int) (*[]*Partition, error) {
	out := make([]*Partition, 0)

	seen := make(map[string]bool)
	queue := list.New()
	queue.PushBack(&queueNode{remaining: radius, partition: p})

	// BFS for surrounding partitons
	for queue.Len() > 0 {
		current := queue.Front()
		node := current.Value.(*queueNode)
		queue.Remove(current)

		// No duplicates
		if _, ok := seen[node.partition.Encoded]; ok {
			continue
		}
		seen[node.partition.Encoded] = true

		out = append(out, node.partition)

		// Identify nearby partitions
		if node.remaining <= 0 {
			continue
		}

		addPartition := func(direction Direction) error {
			partition, err := node.partition.Translate(direction)
			if err != nil {
				return err
			}

			queue.PushBack(&queueNode{remaining: node.remaining - 1, partition: partition})

			return nil
		}

		if err := addPartition(DirUp); err != nil {
			return nil, err
		}
		if err := addPartition(DirDown); err != nil {
			return nil, err
		}
		if err := addPartition(DirLeft); err != nil {
			return nil, err
		}
		if err := addPartition(DirRight); err != nil {
			return nil, err
		}
	}

	return &out, nil
}

func (p *Partition) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Encoded)
}

func (p *Partition) UnmarshalJSON(data []byte) error {
	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	newPartition, err := NewPartitionFromEncoded(encoded)
	if err != nil {
		return err
	}

	p.Chunks = newPartition.Chunks
	p.Encoded = newPartition.Encoded

	return nil
}
