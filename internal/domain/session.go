package domain

import (
	"github.com/google/uuid"
	"math/rand"
	"time"
)

const (
	warmupPercentage = 0.15
	minWarmupSec     = 5
	maxWarmupSec     = 40
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
	calculated := int(float64(targetSec) * warmupPercentage)

	if calculated > maxWarmupSec {
		return maxWarmupSec
	}
	if calculated < minWarmupSec {
		return minWarmupSec
	}
	return calculated
}

func GenerateSteps(targetSec int, r *rand.Rand) []Step {
	warmupCount := warmupStepCount(targetSec, r)
	maxWarmup := maxWarmupDuration(targetSec)

	steps := make([]Step, warmupCount+1)

	for i := 0; i < warmupCount; i++ {
		steps[i] = Step{
			Index:    i,
			Duration: r.Intn(maxWarmup) + 1,
		}
	}

	steps[warmupCount] = Step{
		Index:    warmupCount,
		Duration: targetSec,
	}

	return steps
}

func NewSession(id string, userID string, targetSec int) *Session {
	if id == "" {
		id = uuid.New().String()
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &Session{
		ID:        id,
		UserID:    userID,
		TargetSec: targetSec,
		Steps:     GenerateSteps(targetSec, r),
		StartedAt: time.Now(),
	}
}
