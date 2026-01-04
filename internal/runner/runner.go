package runner

import (
	"context"
	"sync"
	"time"

	"github.com/hperssn/hound/internal/domain"
)

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
		events:  make(chan StepEvent, len(s.Steps)+1), //buffered to avoid blocking runner when sending
	}
}

func (r *sessionRunner) Start() {
	go func() {
		for i := range r.session.Steps {
			r.mu.Lock()
			r.session.CurrentIdx = i
			r.session.Steps[i].StartedAt = time.Now()
			step := r.session.Steps[i]
			r.mu.Unlock()

			ticker := time.NewTicker(1 * time.Second)
			done := time.After(time.Duration(step.Duration) * time.Second)

		stepLoop:
			for {
				select {
				case <-ticker.C:
					elapsed := int(time.Since(step.StartedAt).Seconds())
					select {
					case r.events <- StepEvent{
						Index:   step.Index,
						Elapsed: elapsed,
					}:
					default:
					}

				case <-done:
					ticker.Stop()
					r.mu.Lock()
					r.session.Steps[i].Completed = true
					r.mu.Unlock()
					break stepLoop

				case <-r.ctx.Done():
					ticker.Stop()
					close(r.events)
					return
				}
			}
		}

		r.mu.Lock()
		r.session.Completed = true
		r.mu.Unlock()

		close(r.events)
	}()
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
