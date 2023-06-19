package utils

import (
	"time"

	"github.com/bengosborn/cue/helpers"
)

type Session struct {
	redis *Redis
}

const (
	SessionCookie = "session"
	SessionExpiry = time.Hour
)

func NewSession(redis *Redis) *Session {
	return &Session{redis: redis}
}

// Create a new session
func (s *Session) Create() (string, error) {
	id, err := helpers.GenerateRandomString(32)
	if err != nil {
		return "", err
	}

	return id, nil
}

// Set a new session variable
func (s *Session) Set(id string, key string, value string) error {
	return s.redis.Set(FormatKey(SessionCookie, id, key), value, SessionExpiry)
}

// Get a session variable
func (s *Session) Get(id string, key string) string {
	return s.redis.Get(FormatKey(SessionCookie, id, key))
}
