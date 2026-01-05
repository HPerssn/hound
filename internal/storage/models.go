package storage

import (
	"time"

	"github.com/hperssn/hound/internal/domain"
)

type SuccessLevel string

const (
	SuccessLevelFail  SuccessLevel = "fail"
	SuccessLevelOK    SuccessLevel = "ok"
	SuccessLevelGreat SuccessLevel = "great"
)

type SessionRecord struct {
	ID          string
	UserID      string
	TargetSec   int
	Success     SuccessLevel
	Comment     string
	StartedAt   time.Time
	CompletedAt time.Time
	Steps       []StepRecord
}

type StepRecord struct {
	SessionID string
	Index     int
	Duration  int
	ActualSec int // How long it actually took (might be less if failed)
	Completed bool
}

// FromDomainSession converts a domain.Session to a SessionRecord
func FromDomainSession(s *domain.Session, success SuccessLevel, comment string) *SessionRecord {
	steps := make([]StepRecord, len(s.Steps))
	for i, step := range s.Steps {
		actualSec := 0
		if !step.StartedAt.IsZero() {
			actualSec = int(time.Since(step.StartedAt).Seconds())
		}

		steps[i] = StepRecord{
			SessionID: s.ID,
			Index:     step.Index,
			Duration:  step.Duration,
			ActualSec: actualSec,
			Completed: step.Completed,
		}
	}

	return &SessionRecord{
		ID:          s.ID,
		UserID:      s.UserID,
		TargetSec:   s.TargetSec,
		Success:     success,
		Comment:     comment,
		StartedAt:   s.StartedAt,
		CompletedAt: time.Now(),
		Steps:       steps,
	}
}

