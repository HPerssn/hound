package domain

import (
	"math/rand"
	"testing"
)

func TestMaxWarmupDuration(t *testing.T) {
	tests := []struct {
		name      string
		targetSec int
		expected  int
	}{
		{name: "minimum target", targetSec: 30, expected: 5},
		{name: "medium target", targetSec: 300, expected: 40},
		{name: "large target cap", targetSec: 1200, expected: 40},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maxWarmupDuration(tt.targetSec)

			if result != tt.expected {
				t.Fatalf("maxWarmupDuration(%d) = %d want %d", tt.targetSec, result, tt.expected)
			}
		})
	}
}

func TestGenerateStepsStructure(t *testing.T) {
	targetSec := 300
	r := rand.New(rand.NewSource(1))

	steps := GenerateSteps(targetSec, r)

	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps got %d", len(steps))
	}

	last := steps[len(steps)-1]
	if last.Duration != targetSec {
		t.Fatalf("final step duration = %d, want %d", last.Duration, targetSec)
	}

	maxWarmup := maxWarmupDuration(targetSec)
	for i := 0; i < len(steps)-1; i++ {
		d := steps[i].Duration
		if d < 1 || d > maxWarmup {
			t.Errorf("step %d duration %d out of bounds [1,%d]", i, d, maxWarmup)
		}
	}

	for i, step := range steps {
		if step.Index != i {
			t.Errorf("step index = %d, want %d", step.Index, i)
		}
	}
}

func TestGenerateStepsDeterministic(t *testing.T) {
	targetSec := 300

	r1 := rand.New(rand.NewSource(42))
	r2 := rand.New(rand.NewSource(42))

	steps1 := GenerateSteps(targetSec, r1)
	steps2 := GenerateSteps(targetSec, r2)

	if len(steps1) != len(steps2) {
		t.Fatalf("step count mismatch: %d vs %d", len(steps1), len(steps2))
	}

	for i := range steps1 {
		if steps1[i].Duration != steps2[i].Duration {
			t.Errorf(
				"step %d duration mismatch: %d vs %d",
				i,
				steps1[i].Duration,
				steps2[i].Duration,
			)
		}
	}
}

func TestWarmupStepCountRanges(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	tests := []struct {
		name      string
		targetSec int
		min       int
		max       int
	}{
		{"small target", 100, 5, 6},
		{"medium target", 400, 4, 5},
		{"large target", 700, 3, 4},
		{"very large target", 1200, 3, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := warmupStepCount(tt.targetSec, r)

			if n < tt.min || n > tt.max {
				t.Fatalf(
					"warmupStepCount(%d) = %d, want [%d,%d]",
					tt.targetSec, n, tt.min, tt.max,
				)
			}
		})
	}
}
