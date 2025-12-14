package domain

import "time"

type Session struct {
	ID         string
	userID     string
	TargetSec  int
	steps      []Step
	CurrentIdx int
	StartedAt  time.Time
	Completed  bool
}
