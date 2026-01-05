package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.createTables(); err != nil {
		db.Close()
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteRepository) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		target_sec INTEGER NOT NULL,
		success TEXT NOT NULL,
		comment TEXT,
		started_at DATETIME NOT NULL,
		completed_at DATETIME NOT NULL,
		steps_json TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_completed_at ON sessions(completed_at);
	`

	_, err := r.db.Exec(schema)
	return err
}

func (r *SQLiteRepository) SaveSession(record *SessionRecord) error {
	stepsJSON, err := json.Marshal(record.Steps)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO sessions (id, user_id, target_sec, success, comment, started_at, completed_at, steps_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(
		query,
		record.ID,
		record.UserID,
		record.TargetSec,
		record.Success,
		record.Comment,
		record.StartedAt,
		record.CompletedAt,
		stepsJSON,
	)

	return err
}

func (r *SQLiteRepository) GetSessionsByUser(userID string) ([]SessionRecord, error) {
	query := `
		SELECT id, user_id, target_sec, success, comment, started_at, completed_at, steps_json
		FROM sessions
		WHERE user_id = ?
		ORDER BY completed_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

func (r *SQLiteRepository) GetRecentSessions(userID string, since time.Time) ([]SessionRecord, error) {
	query := `
		SELECT id, user_id, target_sec, success, comment, started_at, completed_at, steps_json
		FROM sessions
		WHERE user_id = ? AND completed_at >= ?
		ORDER BY completed_at DESC
	`

	rows, err := r.db.Query(query, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

func (r *SQLiteRepository) GetSessionStats(userID string) (*SessionStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN success IN ('ok', 'great') THEN 1 ELSE 0 END) as successful,
			AVG(target_sec) as avg_target,
			SUM(target_sec) as total_time
		FROM sessions
		WHERE user_id = ?
	`

	var stats SessionStats
	var totalTime sql.NullInt64
	var avgTarget sql.NullFloat64

	err := r.db.QueryRow(query, userID).Scan(
		&stats.TotalSessions,
		&stats.SuccessfulCount,
		&avgTarget,
		&totalTime,
	)

	if err != nil {
		return nil, err
	}

	if avgTarget.Valid {
		stats.AverageTarget = avgTarget.Float64
	}
	if totalTime.Valid {
		stats.TotalTrainTime = int(totalTime.Int64)
	}
	if stats.TotalSessions > 0 {
		stats.SuccessRate = float64(stats.SuccessfulCount) / float64(stats.TotalSessions) * 100
	}

	return &stats, nil
}

func (r *SQLiteRepository) scanSessions(rows *sql.Rows) ([]SessionRecord, error) {
	var records []SessionRecord

	for rows.Next() {
		var record SessionRecord
		var stepsJSON string

		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.TargetSec,
			&record.Success,
			&record.Comment,
			&record.StartedAt,
			&record.CompletedAt,
			&stepsJSON,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(stepsJSON), &record.Steps); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
