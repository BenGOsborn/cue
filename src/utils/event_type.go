package utils

type EventType int

const (
	// General events
	Error EventType = iota

	// Proximity service events
	ProximitySendLocation
	ProximityRequestNearby
)
