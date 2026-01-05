package storage

import "time"

type Repository interface {
	SaveSession(record *SessionRecord) error

	GetSessionsByUser(userID string) ([]SessionRecord, error)

	GetRecentSessions(userID string, since time.Time) ([]SessionRecord, error)

	GetSessionStats(userID string) (*SessionStats, error)

	Close() error
}

type SessionStats struct {
	TotalSessions   int     `json:"totalSessions"`
	SuccessfulCount int     `json:"successfulCount"`
	AverageTarget   float64 `json:"averageTarget"`
	TotalTrainTime  int     `json:"totalTrainTime"`
	SuccessRate     float64 `json:"successRate"`
}
