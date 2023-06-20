package utils

type BrokerMessage struct {
	Id        string    `json:"id"`
	Receiver  string    `json:"receiver"`
	User      string    `json:"user"`
	EventType EventType `json:"eventType"`
	Body      string    `json:"body"`
}
