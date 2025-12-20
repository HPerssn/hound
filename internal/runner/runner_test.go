package runner_test

import (
	"testing"
	"time"

	"github.com/hperssn/hound/internal/domain"
	"github.com/hperssn/hound/internal/runner"
)

func TestSessionRunner_Start(t *testing.T) {
	// Create a short session
	s := &domain.Session{
		Steps: []domain.Step{
			{Index: 0, Duration: 1},
			{Index: 1, Duration: 1},
		},
	}

	r := runner.NewSessionRunner(s)
	r.Start()

	completedSteps := 0
	for step := range r.Events() { // assume you add a public getter for events
		if step.Index != completedSteps {
			t.Errorf("expected step %d, got %d", completedSteps, step.Index)
		}
		completedSteps++
		if completedSteps == len(s.Steps) {
			break
		}
	}

	// Give goroutine a tiny moment to update session.Completed
	time.Sleep(10 * time.Millisecond)
	if !r.Session().Completed {
		t.Errorf("expected session to be completed")
	}
}

func TestSessionRunner_Stop(t *testing.T) {
	s := &domain.Session{
		Steps: []domain.Step{
			{Index: 0, Duration: 10}, // long step
		},
	}

	r := runner.NewSessionRunner(s)
	r.Start()
	r.Stop()

	// Allow goroutine to finish
	time.Sleep(10 * time.Millisecond)

	if r.Session().Completed {
		t.Errorf("session should not be marked completed after Stop()")
	}
}
