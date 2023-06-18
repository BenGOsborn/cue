package gateway

type QueueMessage struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Body string `json:"body"`
}
