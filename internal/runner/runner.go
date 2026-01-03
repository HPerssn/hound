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

	events chan domain.Step
}

func NewSessionRunner(s *domain.Session) *sessionRunner {
	ctx, cancel := context.WithCancel(context.Background())
	return &sessionRunner{
		session: s,
		ctx:     ctx,
		cancel:  cancel,
		events:  make(chan domain.Step, len(s.Steps)), //buffered to avoid blocking runner when sending
	}
}

func (r *sessionRunner) Start() {
	go func() {
		for i, step := range r.session.Steps {

			select {
			case <-time.After(time.Duration(step.Duration) * time.Second):
				r.mu.Lock()
				r.session.Steps[i].Completed = true
				r.mu.Unlock()

				r.mu.Lock()
				updated := r.session.Steps[i]
				r.mu.Unlock()

				r.events <- updated

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

func (r *sessionRunner) Events() <-chan domain.Step {
	return r.events

}

func (r *sessionRunner) Session() *domain.Session {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *r.session
	return &copy
}
