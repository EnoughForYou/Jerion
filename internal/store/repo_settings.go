package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrSettingNotFound = errors.New("setting not found")

// SetSetting writes or updates a configuration key-value pair.
func (s *Store) SetSetting(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
		INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at;
	`
	_, err := s.db.Exec(query, key, value, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("set setting %s: %w", key, err)
	}
	return nil
}

// GetSetting retrieves a configuration value by key.
func (s *Store) GetSetting(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var val string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrSettingNotFound
		}
		return "", fmt.Errorf("get setting %s: %w", key, err)
	}
	return val, nil
}

// GetAllSettings returns all saved key-value configuration items.
func (s *Store) GetAllSettings() (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}
	defer rows.Close()

	res := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan setting row: %w", err)
		}
		res[k] = v
	}
	return res, rows.Err()
}

// DeleteSetting removes a setting key.
func (s *Store) DeleteSetting(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM settings WHERE key = ?", key)
	if err != nil {
		return fmt.Errorf("delete setting %s: %w", key, err)
	}
	return nil
}
