package store

import (
	"database/sql"
	"fmt"
	"time"
)

type migration struct {
	version int
	name    string
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		name:    "initial_schema",
		sql: `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS devices (
    serial TEXT PRIMARY KEY,
    model TEXT NOT NULL,
    manufacturer TEXT NOT NULL,
    android_version TEXT NOT NULL,
    sdk INTEGER NOT NULL,
    security_patch TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS scan_results (
    id TEXT PRIMARY KEY,
    device_serial TEXT NOT NULL,
    mode TEXT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP NOT NULL,
    rules_version TEXT NOT NULL,
    app_version TEXT NOT NULL,
    partial INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS scan_findings (
    id TEXT PRIMARY KEY,
    scan_id TEXT NOT NULL REFERENCES scan_results(id) ON DELETE CASCADE,
    rule_id TEXT NOT NULL,
    package_name TEXT NOT NULL,
    category TEXT NOT NULL,
    severity TEXT NOT NULL,
    confidence TEXT NOT NULL,
    weight INTEGER NOT NULL,
    title TEXT NOT NULL,
    why_flagged TEXT NOT NULL,
    fp_hint TEXT NOT NULL,
    evidence_json TEXT NOT NULL,
    actions_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_findings_scan_id ON scan_findings(scan_id);
CREATE INDEX IF NOT EXISTS idx_findings_package ON scan_findings(package_name);

CREATE TABLE IF NOT EXISTS actions_log (
    id TEXT PRIMARY KEY,
    device_serial TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    action_type TEXT NOT NULL,
    package_name TEXT NOT NULL,
    details_json TEXT NOT NULL,
    reversible INTEGER NOT NULL DEFAULT 0,
    undone INTEGER NOT NULL DEFAULT 0,
    undone_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_actions_device ON actions_log(device_serial);

CREATE TABLE IF NOT EXISTS allowlist (
    id TEXT PRIMARY KEY,
    package_name TEXT NOT NULL,
    cert_sha256 TEXT NOT NULL,
    reason TEXT NOT NULL,
    added_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_allowlist_package ON allowlist(package_name);

CREATE TABLE IF NOT EXISTS intel_hashes (
    sha256 TEXT PRIMARY KEY,
    classification TEXT NOT NULL,
    threat_name TEXT NOT NULL,
    source TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
`,
	},
	{
		version: 2,
		name:    "add_total_apps_to_scan_results",
		sql: `
ALTER TABLE scan_results ADD COLUMN total_apps INTEGER NOT NULL DEFAULT 0;
`,
	},
}

// RunMigrations executes all unapplied schema migrations in sequence.
func (s *Store) RunMigrations() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure migration table exists
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	for _, m := range migrations {
		var exists int
		err := s.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if exists > 0 {
			continue
		}

		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec(m.sql); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute migration %d (%s): %w", m.version, m.name, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
			m.version, m.name, time.Now().UTC(),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
	}

	return nil
}
