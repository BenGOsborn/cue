package gateway

import (
	"encoding/json"
	"time"

	"github.com/bengosborn/cue/helpers"
	"github.com/bengosborn/cue/utils"
)

type Session struct {
	redis *utils.Redis
}

type SessionData struct {
	Token     string `json:"token"`
	CSRFToken string `json:"csrfToken"`
}

const (
	SessionCookie = "session"
	SessionExpiry = time.Hour
)

func NewSession(redis *utils.Redis) *Session {
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

// Set the session data
func (s *Session) Set(id string, value *SessionData) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.redis.Set(utils.FormatKey(SessionCookie, id), string(data), SessionExpiry)
}

// Retrieve a session
func (s *Session) Get(id string) (*SessionData, error) {
	raw := s.redis.Get(utils.FormatKey(SessionCookie, id))

	data := SessionData{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// Delete a session
func (s *Session) Delete(id string) error {
	return s.redis.Remove(utils.FormatKey(SessionCookie, id))
}