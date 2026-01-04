package runner_test

import (
	"testing"
	"time"

	"github.com/hperssn/hound/internal/domain"
	"github.com/hperssn/hound/internal/runner"
)

func TestSessionRunner_StartStep(t *testing.T) {
	s := &domain.Session{
		ID: "test-session",
		Steps: []domain.Step{
			{Index: 0, Duration: 2},
			{Index: 1, Duration: 2},
		},
	}

	r := runner.NewSessionRunner(s)

	// Start first step
	if err := r.StartStep(0); err != nil {
		t.Fatalf("failed to start step 0: %v", err)
	}

	// Wait for first step to complete
	time.Sleep(3 * time.Second)

	sess := r.Session()
	if !sess.Steps[0].Completed {
		t.Error("step 0 should be completed")
	}

	// Start second step
	if err := r.StartStep(1); err != nil {
		t.Fatalf("failed to start step 1: %v", err)
	}

	// Wait for second step to complete
	time.Sleep(3 * time.Second)

	sess = r.Session()
	if !sess.Steps[1].Completed {
		t.Error("step 1 should be completed")
	}

	if !sess.Completed {
		t.Error("session should be marked completed after all steps")
	}
}

func TestSessionRunner_StopStep(t *testing.T) {
	s := &domain.Session{
		ID: "test-session",
		Steps: []domain.Step{
			{Index: 0, Duration: 10},
		},
	}

	r := runner.NewSessionRunner(s)

	// Start the step
	if err := r.StartStep(0); err != nil {
		t.Fatalf("failed to start step: %v", err)
	}

	// Let it run for a bit
	time.Sleep(2 * time.Second)

	// Stop the step
	if err := r.StopStep(0); err != nil {
		t.Fatalf("failed to stop step: %v", err)
	}

	sess := r.Session()
	if sess.Steps[0].Completed {
		t.Error("step should not be completed after stopping")
	}
}

func TestSessionRunner_Events(t *testing.T) {
	s := &domain.Session{
		ID: "test-session",
		Steps: []domain.Step{
			{Index: 0, Duration: 2},
		},
	}

	r := runner.NewSessionRunner(s)
	events := r.Events()

	if err := r.StartStep(0); err != nil {
		t.Fatalf("failed to start step: %v", err)
	}

	// Collect events for 3 seconds
	receivedEvents := 0
	timeout := time.After(3 * time.Second)

	for {
		select {
		case event := <-events:
			if event.Index != 0 {
				t.Errorf("expected event for step 0, got step %d", event.Index)
			}
			receivedEvents++
		case <-timeout:
			if receivedEvents < 1 {
				t.Error("should have received at least one event")
			}
			return
		}
	}
}

func TestSessionManager_StartSession(t *testing.T) {
	manager := runner.NewSessionManager()
	s := domain.NewSession("", "", 10)

	if err := manager.StartSession(s); err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	retrieved, ok := manager.GetSession(s.ID)
	if !ok {
		t.Fatal("session not found after starting")
	}

	if retrieved.ID != s.ID {
		t.Errorf("expected session ID %s, got %s", s.ID, retrieved.ID)
	}
}

func TestSessionManager_StopSession(t *testing.T) {
	manager := runner.NewSessionManager()
	s := domain.NewSession("", "", 10)

	if err := manager.StartSession(s); err != nil {
		t.Fatal(err)
	}

	if err := manager.StopSession(s.ID); err != nil {
		t.Fatalf("failed to stop session: %v", err)
	}

	// After stopping, session should be removed from manager
	_, ok := manager.GetSession(s.ID)
	if ok {
		t.Error("session should be removed after stopping")
	}
}
