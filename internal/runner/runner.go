package runner

import (
	"context"
	"sync"
	"time"

	"github.com/hperssn/hound/internal/domain"
)

type StepEvent struct {
	Index     int  `json:"index"`
	Duration  int  `json:"duration"`
	Elapsed   int  `json:"elapsed"`
	Completed bool `json:"completed"`
}

type sessionRunner struct {
	mu sync.Mutex

	session *domain.Session

	ctx    context.Context
	cancel context.CancelFunc

	events chan StepEvent
}

func NewSessionRunner(s *domain.Session) *sessionRunner {
	ctx, cancel := context.WithCancel(context.Background())
	return &sessionRunner{
		session: s,
		ctx:     ctx,
		cancel:  cancel,
		events:  make(chan StepEvent, len(s.Steps)), //buffered to avoid blocking runner when sending
	}
}

func (r *sessionRunner) Start() {
	go func() {
		for i := range r.session.Steps {

			r.mu.Lock()
			r.session.CurrentIdx = i
			r.session.Steps[i].StartedAt = time.Now()
			started := r.session.Steps[i].StartedAt
			r.mu.Unlock()

			select {
			case <-time.After(time.Duration(r.session.Steps[i].Duration) * time.Second):
				r.mu.Lock()
				r.session.Steps[i].Completed = true
				step := r.session.Steps[i]
				r.mu.Unlock()

				elapsed := int(started.Sub(r.session.StartedAt).Seconds())

				r.events <- StepEvent{
					Index:     step.Index,
					Duration:  step.Duration,
					Elapsed:   elapsed,
					Completed: step.Completed,
				}
			case <-r.ctx.Done():
				return
			}
		}
		r.mu.Lock()
		r.session.Completed = true
		r.mu.Unlock()
	}()
}

func (r *sessionRunner) Stop() {
	r.cancel()
	close(r.events)
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
