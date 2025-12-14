package domain

import "time"

type Step struct {
	Index     int
	Duration  int
	StartedAt time.Time
	Completed bool
}
