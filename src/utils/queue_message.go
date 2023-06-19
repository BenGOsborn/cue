package utils

type QueueMessage struct {
	Receiver  string    `json:"receiver"`
	User      string    `json:"user"`
	EventType EventType `json:"eventType"`
	Body      string    `json:"body"`
}
