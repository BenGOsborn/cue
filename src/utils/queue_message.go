package utils

import "github.com/coreos/go-oidc/v3/oidc"

type QueueMessage struct {
	Receiver  string        `json:"receiver"`
	User      *oidc.IDToken `json:"user"`
	EventType EventType     `json:"eventType"`
	Body      string        `json:"body"`
}
