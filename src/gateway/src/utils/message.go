package gateway

import (
	"github.com/bengosborn/cue/utils"
)

type Message struct {
	Auth      string          `json:"auth"`
	EventType utils.EventType `json:"eventType"`
	Body      string          `json:"body"`
}
