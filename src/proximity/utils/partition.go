package utils

type Partition struct {
	latitude  float32
	longitude float32
	depth     int
	encoded   string
}

// Create a new partition from a latitude and longitude
func NewPartition() (*Partition, error) {
	return nil, nil
}

// Format the partition
func (p *Partition) String() string {
	return p.encoded
}
