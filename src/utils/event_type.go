package utils

type EventType int

const (
	// Proximity service events
	ProximityError EventType = iota
	ProximitySendLocation
	ProximityRequestNearby
)
