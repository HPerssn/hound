package domain

import (
	"github.com/google/uuid"
	"math/rand"
	"time"
)

type Session struct {
	ID         string
	UserID     string
	TargetSec  int
	Steps      []Step
	CurrentIdx int
	StartedAt  time.Time
	Completed  bool
}

func warmupStepCount(targetSec int, r *rand.Rand) int {
	switch {
	case targetSec < 240:
		return 5 + r.Intn(2)
	case targetSec < 600:
		return 4 + r.Intn(2)
	case targetSec < 900:
		return 3 + r.Intn(2)
	default:
		return 2 + r.Intn(2)
	}
}

func maxWarmupDuration(targetSec int) int {
	max := int(float64(targetSec) * 0.15)

	if max > 40 {
		return 40
	}
	if max < 5 {
		return 5
	}
	return max
}

func GenerateSteps(targetSec int, r *rand.Rand) []Step {
	warmups := warmupStepCount(targetSec, r)
	maxWarmup := maxWarmupDuration(targetSec)

	steps := make([]Step, warmups+1)

	for i := 0; i < warmups; i++ {
		steps[i] = Step{
			Index:     i,
			Duration:  r.Intn(maxWarmup) + 1,
			Completed: false,
		}
	}

	steps[warmups] = Step{
		Index:     warmups,
		Duration:  targetSec,
		Completed: false,
	}

	return steps
}

func NewSession(id string, userID string, targetSec int) *Session {
	if id == "" {
		id = uuid.New().String()
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &Session{
		ID:         id,
		UserID:     userID,
		TargetSec:  targetSec,
		Steps:      GenerateSteps(targetSec, r),
		CurrentIdx: 0,
		StartedAt:  time.Now(),
		Completed:  false,
	}
}
