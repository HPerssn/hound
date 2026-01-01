package runner_test

import (
	"testing"
	"time"

	"github.com/hperssn/hound/internal/domain"
	"github.com/hperssn/hound/internal/runner"
)

func TestSessionManager_StartAndGet(t *testing.T) {
	m := runner.NewSessionManager()

	s := &domain.Session{
		ID: "session-1",
		Steps: []domain.Step{
			{Index: 0, Duration: 1},
		},
	}

	err := m.StartSession(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := m.GetSession(s.ID)
	if !ok {
		t.Fatalf("expected session to exist")
	}
	if got.ID != s.ID {
		t.Fatalf("expected session ID %s, got %s", s.ID, got.ID)
	}
}

func TestSessionManager_DuplicateStart(t *testing.T) {
	m := runner.NewSessionManager()

	s := &domain.Session{
		ID: "session-dup",
		Steps: []domain.Step{
			{Index: 0, Duration: 1},
		},
	}
	if err := m.StartSession(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.StartSession(s); err == nil {
		t.Fatalf("Expected error on duplicate start")
	}
}

func TestSessionManager_Stop(t *testing.T) {
	m := runner.NewSessionManager()

	s := &domain.Session{
		ID: "session-stop",
		Steps: []domain.Step{
			{Index: 0, Duration: 10},
		},
	}
	if err := m.StartSession(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.StopSession(s.ID); err != nil {
		t.Fatalf("unexpected error stopping session: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	sess, ok := m.GetSession(s.ID)
	if !ok {
		t.Fatalf("expected session to still exist")
	}
	if sess.Completed {
		t.Fatalf("stopped session should not be ompleted")
	}
}

func TestSessionManager_StopMissing(t *testing.T) {
	m := runner.NewSessionManager()

	if err := m.StopSession("missing"); err == nil {
		t.Fatalf("expected error when stopping missing session")
	}
}
