package runner

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/hperssn/hound/internal/domain"
)

type stepControl struct {
	step         *domain.Step
	cancel       chan struct{}
	paused       bool
	elapsedSoFar int
}

type sessionRunner struct {
	mu sync.Mutex

	session *domain.Session
	ctx     context.Context
	cancel  context.CancelFunc

	events chan StepEvent
	steps  map[int]*stepControl
}

func NewSessionRunner(s *domain.Session) *sessionRunner {
	ctx, cancel := context.WithCancel(context.Background())
	return &sessionRunner{
		session: s,
		ctx:     ctx,
		cancel:  cancel,
		events:  make(chan StepEvent, len(s.Steps)+1),
		steps:   make(map[int]*stepControl),
	}
}

func (r *sessionRunner) StartStep(idx int) error {
	r.mu.Lock()
	if idx < 0 || idx >= len(r.session.Steps) {
		r.mu.Unlock()
		return errors.New("invalid step index")
	}

	step := &r.session.Steps[idx]

	sc, exists := r.steps[idx]
	if !exists {
		sc = &stepControl{
			step:         step,
			cancel:       make(chan struct{}),
			elapsedSoFar: 0,
		}
		r.steps[idx] = sc
	} else if !sc.paused && !step.Completed && !step.StartedAt.IsZero() {
		r.mu.Unlock()
		return errors.New("step already running")
	} else if sc.paused {
		sc.cancel = make(chan struct{})
		sc.paused = false
	}

	startTime := time.Now()
	if step.StartedAt.IsZero() {
		step.StartedAt = startTime
	}

	r.session.CurrentIdx = idx
	r.mu.Unlock()

	go func(s *domain.Step, sc *stepControl) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				elapsed := sc.elapsedSoFar + int(time.Since(startTime).Seconds())
				select {
				case r.events <- StepEvent{Index: s.Index, Elapsed: elapsed}:
				default:
				}
				if elapsed >= s.Duration {
					r.mu.Lock()
					s.Completed = true
					r.mu.Unlock()
					r.sendStepDone()
					return
				}

			case <-sc.cancel:
				r.mu.Lock()
				sc.elapsedSoFar += int(time.Since(startTime).Seconds())
				r.mu.Unlock()
				return

			case <-r.ctx.Done():
				return
			}
		}
	}(step, sc)

	return nil
}

func (r *sessionRunner) StopStep(idx int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sc, ok := r.steps[idx]
	if !ok || sc.paused {
		return errors.New("step not running")
	}

	sc.paused = true
	close(sc.cancel)
	return nil
}

func (r *sessionRunner) Stop() {
	r.cancel()
}

func (r *sessionRunner) Events() <-chan StepEvent {
	return r.events
}

func (r *sessionRunner) Session() *domain.Session {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *r.session
	return &copy
}

func (r *sessionRunner) sendStepDone() {
	r.mu.Lock()
	defer r.mu.Unlock()

	allDone := true
	for _, st := range r.session.Steps {
		if !st.Completed {
			allDone = false
			break
		}
	}
	if allDone {
		r.session.Completed = true
		select {
		case r.events <- StepEvent{}: // empty event indicates session done
		default:
		}
	}
}
