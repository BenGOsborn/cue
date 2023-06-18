package gateway

type Message struct {
	Id   string  `json:"id"`
	Auth *string `json:"auth,omitempty"`
	Type string  `json:"type"`
	Body string  `json:"body"`
}
