package webauthn

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	platformerrors "github.com/primandproper/platform-go/v5/errors"

	"github.com/go-webauthn/webauthn/webauthn"
)

// SessionStore persists WebAuthn session data for the duration of the ceremony.
type SessionStore interface {
	SaveSession(ctx context.Context, challenge string, session *webauthn.SessionData, ttl time.Duration) error
	GetSession(ctx context.Context, challenge string) (*webauthn.SessionData, error)
}

// InMemorySessionStore is an in-memory session store. Suitable for single-instance deployments only.
type InMemorySessionStore struct {
	sessions map[string]*sessionEntry
	mu       sync.RWMutex
}

type sessionEntry struct {
	expiresAt time.Time
	data      []byte
}

// NewInMemorySessionStore creates a new in-memory session store.
func NewInMemorySessionStore() *InMemorySessionStore {
	s := &InMemorySessionStore{sessions: make(map[string]*sessionEntry)}
	go s.cleanupLoop()
	return s
}

func (s *InMemorySessionStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()

		for k, v := range s.sessions {
			if v.expiresAt.Before(now) {
				delete(s.sessions, k)
			}
		}

		s.mu.Unlock()
	}
}

// SaveSession stores session data keyed by challenge.
func (s *InMemorySessionStore) SaveSession(ctx context.Context, challenge string, session *webauthn.SessionData, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[challenge] = &sessionEntry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// GetSession retrieves session data by challenge.
func (s *InMemorySessionStore) GetSession(ctx context.Context, challenge string) (*webauthn.SessionData, error) {
	s.mu.RLock()
	entry, ok := s.sessions[challenge]
	s.mu.RUnlock()

	if !ok || entry == nil {
		return nil, platformerrors.New("session not found")
	}

	if time.Now().After(entry.expiresAt) {
		s.mu.Lock()
		delete(s.sessions, challenge)
		s.mu.Unlock()
		return nil, platformerrors.New("session expired")
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(entry.data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}
