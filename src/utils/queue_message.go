package gateway

type QueueMessage struct {
	Receiver string `json:"id"`
	Type     string `json:"type"`
	Body     string `json:"body"`
}
