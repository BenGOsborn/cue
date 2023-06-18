package gateway

type Message struct {
	Auth *string `json:"auth,omitempty"`
	Type string  `json:"type"`
	Body string  `json:"body"`
}
