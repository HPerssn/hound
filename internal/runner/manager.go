package runner

import (
	"errors"
	"sync"
	"time"

	"github.com/hperssn/hound/internal/domain"
)

var (
	ErrSessionExists   = errors.New("session already exists")
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidStep     = errors.New("invalid step index")
)

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]*sessionRunner
}

func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions: make(map[string]*sessionRunner),
	}

	go m.cleanupLoop()

	return m
}

func (m *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanupOldSessions()
	}
}

func (m *SessionManager) cleanupOldSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)

	for id, runner := range m.sessions {
		sess := runner.Session()
		if sess.Completed && sess.StartedAt.Before(cutoff) {
			runner.Stop()
			delete(m.sessions, id)
		}
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

	return nil
}

func (m *SessionManager) StopSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.sessions[id]
	if !exists {
		return ErrSessionNotFound
	}

	r.Stop()
	delete(m.sessions, id)
	return nil
}

func (m *SessionManager) CompleteSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.sessions[id]
	if !exists {
		return ErrSessionNotFound
	}

	r.MarkCompleted()
	r.StopAllSteps()

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

func (m *SessionManager) StartStep(sessionID string, idx int) error {
	m.mu.Lock()
	r, exists := m.sessions[sessionID]
	m.mu.Unlock()

	if !exists {
		return ErrSessionNotFound
	}

	sess := r.Session()
	if idx < 0 || idx >= len(sess.Steps) {
		return ErrInvalidStep
	}

	return r.StartStep(idx)
}

func (m *SessionManager) StopStep(sessionID string, idx int) error {
	m.mu.Lock()
	r, exists := m.sessions[sessionID]
	m.mu.Unlock()

	if !exists {
		return ErrSessionNotFound
	}

	sess := r.Session()
	if idx < 0 || idx >= len(sess.Steps) {
		return ErrInvalidStep
	}

	return r.StopStep(idx)
}
