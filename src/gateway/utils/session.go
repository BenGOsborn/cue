package gateway

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bengosborn/cue/helpers"
	"github.com/redis/go-redis/v9"
)

type Session struct {
	ctx   context.Context
	redis *redis.Client
}

type SessionData struct {
	Token     string `json:"token"`
	CSRFToken string `json:"csrfToken"`
}

const (
	SessionCookie = "session"
	SessionExpiry = time.Hour
)

// Create a new session
func NewSession(ctx context.Context, redis *redis.Client) *Session {
	return &Session{redis: redis, ctx: ctx}
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

	return s.redis.Set(s.ctx, helpers.FormatKey(SessionCookie, id), string(data), SessionExpiry).Err()
}

// Retrieve a session
func (s *Session) Get(id string) (*SessionData, error) {
	raw := s.redis.Get(s.ctx, helpers.FormatKey(SessionCookie, id)).Val()

	data := SessionData{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// Delete a session
func (s *Session) Delete(id string) error {
	return s.redis.Del(s.ctx, helpers.FormatKey(SessionCookie, id)).Err()
}
