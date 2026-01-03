package runner

import (
	"errors"
	"sync"

	"github.com/hperssn/hound/internal/domain"
)

var (
	ErrSessionExists   = errors.New("session already exists")
	ErrSessionNotFound = errors.New("session not found")
)

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]*sessionRunner
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*sessionRunner),
	}
}

func (m *SessionManager) Events(id string) (<-chan StepEvent, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.sessions[id]
	if !ok {
		return nil, false
	}

	return r.Events(), true
}

func (m *SessionManager) StartSession(s *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[s.ID]; exists {
		return ErrSessionExists
	}

	r := NewSessionRunner(s)
	m.sessions[s.ID] = r
	r.Start()

	return nil
}

func (m *SessionManager) StopSession(id string) error {

	m.mu.Lock()
	r, exists := m.sessions[id]
	m.mu.Unlock()

	if !exists {
		return ErrSessionNotFound
	}

	r.Stop()
	return nil
}

func (m *SessionManager) GetSession(id string) (*domain.Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.sessions[id]
	if !exists {
		return nil, false
	}
	return r.Session(), true
}
