package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrDeviceNotFound = errors.New("device not found in store")

// DeviceRecord stores persisted metadata for connected Android devices.
type DeviceRecord struct {
	Serial         string    `json:"serial"`
	Model          string    `json:"model"`
	Manufacturer   string    `json:"manufacturer"`
	AndroidVersion string    `json:"androidVersion"`
	SDK            int       `json:"sdk"`
	SecurityPatch  string    `json:"securityPatch"`
	Fingerprint    string    `json:"fingerprint"`
	FirstSeen      time.Time `json:"firstSeen"`
	LastSeen       time.Time `json:"lastSeen"`
}

// UpsertDevice records a device connection and updates last seen timestamp.
func (s *Store) UpsertDevice(d DeviceRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if d.FirstSeen.IsZero() {
		d.FirstSeen = now
	}
	d.LastSeen = now

	query := `
		INSERT INTO devices (serial, model, manufacturer, android_version, sdk, security_patch, fingerprint, first_seen, last_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(serial) DO UPDATE SET
			model = excluded.model,
			manufacturer = excluded.manufacturer,
			android_version = excluded.android_version,
			sdk = excluded.sdk,
			security_patch = excluded.security_patch,
			fingerprint = excluded.fingerprint,
			last_seen = excluded.last_seen;
	`
	_, err := s.db.Exec(query,
		d.Serial, d.Model, d.Manufacturer, d.AndroidVersion, d.SDK, d.SecurityPatch, d.Fingerprint, d.FirstSeen, d.LastSeen,
	)
	if err != nil {
		return fmt.Errorf("upsert device %s: %w", d.Serial, err)
	}
	return nil
}

// GetDevice retrieves a device by serial number.
func (s *Store) GetDevice(serial string) (*DeviceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var d DeviceRecord
	query := `
		SELECT serial, model, manufacturer, android_version, sdk, security_patch, fingerprint, first_seen, last_seen
		FROM devices WHERE serial = ?
	`
	err := s.db.QueryRow(query, serial).Scan(
		&d.Serial, &d.Model, &d.Manufacturer, &d.AndroidVersion, &d.SDK, &d.SecurityPatch, &d.Fingerprint, &d.FirstSeen, &d.LastSeen,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("get device %s: %w", serial, err)
	}
	return &d, nil
}

// ListDevices returns all devices ever connected.
func (s *Store) ListDevices() ([]DeviceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT serial, model, manufacturer, android_version, sdk, security_patch, fingerprint, first_seen, last_seen
		FROM devices ORDER BY last_seen DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []DeviceRecord
	for rows.Next() {
		var d DeviceRecord
		err := rows.Scan(
			&d.Serial, &d.Model, &d.Manufacturer, &d.AndroidVersion, &d.SDK, &d.SecurityPatch, &d.Fingerprint, &d.FirstSeen, &d.LastSeen,
		)
		if err != nil {
			return nil, fmt.Errorf("scan device row: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}
