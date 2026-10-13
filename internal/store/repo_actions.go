package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrActionNotFound = errors.New("action log entry not found")

// ActionLogRecord tracks remediation and modification actions for auditing and undo operations.
type ActionLogRecord struct {
	ID           string         `json:"id"`
	DeviceSerial string         `json:"deviceSerial"`
	Timestamp    time.Time      `json:"timestamp"`
	ActionType   string         `json:"actionType"`
	Package      string         `json:"package"`
	Details      map[string]any `json:"details"`
	Reversible   bool           `json:"reversible"`
	Undone       bool           `json:"undone"`
	UndoneAt     *time.Time     `json:"undoneAt,omitempty"`
}

// LogAction persists an action in the audit log.
func (s *Store) LogAction(act ActionLogRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if act.Timestamp.IsZero() {
		act.Timestamp = time.Now().UTC()
	}

	detailsJSON, err := json.Marshal(act.Details)
	if err != nil {
		return fmt.Errorf("marshal action details: %w", err)
	}

	revInt := 0
	if act.Reversible {
		revInt = 1
	}
	undoneInt := 0
	if act.Undone {
		undoneInt = 1
	}

	query := `
		INSERT INTO actions_log (
			id, device_serial, timestamp, action_type, package_name, details_json, reversible, undone, undone_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = s.db.Exec(query,
		act.ID, act.DeviceSerial, act.Timestamp, act.ActionType, act.Package, string(detailsJSON),
		revInt, undoneInt, act.UndoneAt,
	)
	if err != nil {
		return fmt.Errorf("insert action log %s: %w", act.ID, err)
	}
	return nil
}

// GetAction retrieves an individual action entry.
func (s *Store) GetAction(id string) (*ActionLogRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var act ActionLogRecord
	var detailsStr string
	var revInt, undoneInt int
	var undoneAt sql.NullTime

	query := `
		SELECT id, device_serial, timestamp, action_type, package_name, details_json, reversible, undone, undone_at
		FROM actions_log WHERE id = ?
	`
	err := s.db.QueryRow(query, id).Scan(
		&act.ID, &act.DeviceSerial, &act.Timestamp, &act.ActionType, &act.Package, &detailsStr,
		&revInt, &undoneInt, &undoneAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrActionNotFound
		}
		return nil, fmt.Errorf("get action %s: %w", id, err)
	}

	act.Reversible = revInt == 1
	act.Undone = undoneInt == 1
	if undoneAt.Valid {
		act.UndoneAt = &undoneAt.Time
	}

	if len(detailsStr) > 0 {
		_ = json.Unmarshal([]byte(detailsStr), &act.Details)
	}

	return &act, nil
}

// MarkActionUndone updates an action record to flag that it was successfully reversed.
func (s *Store) MarkActionUndone(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	query := "UPDATE actions_log SET undone = 1, undone_at = ? WHERE id = ?"
	res, err := s.db.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("mark action undone %s: %w", id, err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrActionNotFound
	}
	return nil
}

// ListActionLogs returns actions sorted chronologically descending.
func (s *Store) ListActionLogs(deviceSerial string) ([]ActionLogRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT id, device_serial, timestamp, action_type, package_name, details_json, reversible, undone, undone_at
		FROM actions_log
	`
	var args []any
	if len(deviceSerial) > 0 {
		query += " WHERE device_serial = ? "
		args = append(args, deviceSerial)
	}
	query += " ORDER BY timestamp DESC "

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list action logs: %w", err)
	}
	defer rows.Close()

	var logs []ActionLogRecord
	for rows.Next() {
		var act ActionLogRecord
		var detailsStr string
		var revInt, undoneInt int
		var undoneAt sql.NullTime

		err := rows.Scan(
			&act.ID, &act.DeviceSerial, &act.Timestamp, &act.ActionType, &act.Package, &detailsStr,
			&revInt, &undoneInt, &undoneAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan action log row: %w", err)
		}

		act.Reversible = revInt == 1
		act.Undone = undoneInt == 1
		if undoneAt.Valid {
			act.UndoneAt = &undoneAt.Time
		}

		if len(detailsStr) > 0 {
			_ = json.Unmarshal([]byte(detailsStr), &act.Details)
		}

		logs = append(logs, act)
	}
	return logs, rows.Err()
}
