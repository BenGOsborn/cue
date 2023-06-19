package gateway

import (
	"github.com/bengosborn/cue/utils"
)

type Message struct {
	SessionId string          `json:"sessionId"`
	EventType utils.EventType `json:"eventType"`
	Body      string          `json:"body"`
}
