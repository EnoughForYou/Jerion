package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrIntelNotFound = errors.New("hash not found in local threat intelligence")

// IntelHashRecord represents a known file hash signature.
type IntelHashRecord struct {
	SHA256         string    `json:"sha256"`
	Classification string    `json:"classification"`
	ThreatName     string    `json:"threatName"`
	Source         string    `json:"source"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// UpsertIntelHash inserts or updates a threat hash record.
func (s *Store) UpsertIntelHash(h IntelHashRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	h.SHA256 = strings.ToLower(strings.TrimSpace(h.SHA256))
	if h.UpdatedAt.IsZero() {
		h.UpdatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO intel_hashes (sha256, classification, threat_name, source, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(sha256) DO UPDATE SET
			classification = excluded.classification,
			threat_name = excluded.threat_name,
			source = excluded.source,
			updated_at = excluded.updated_at;
	`
	_, err := s.db.Exec(query, h.SHA256, h.Classification, h.ThreatName, h.Source, h.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert intel hash %s: %w", h.SHA256, err)
	}
	return nil
}

// LookupIntelHash finds a hash signature in the local database.
func (s *Store) LookupIntelHash(sha256 string) (*IntelHashRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	norm := strings.ToLower(strings.TrimSpace(sha256))
	var h IntelHashRecord
	query := "SELECT sha256, classification, threat_name, source, updated_at FROM intel_hashes WHERE sha256 = ?"
	err := s.db.QueryRow(query, norm).Scan(&h.SHA256, &h.Classification, &h.ThreatName, &h.Source, &h.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrIntelNotFound
		}
		return nil, fmt.Errorf("lookup intel hash %s: %w", sha256, err)
	}
	return &h, nil
}

// ImportIntelHashes performs a bulk import within a single transaction.
func (s *Store) ImportIntelHashes(hashes []IntelHashRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin import tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO intel_hashes (sha256, classification, threat_name, source, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(sha256) DO UPDATE SET
			classification = excluded.classification,
			threat_name = excluded.threat_name,
			source = excluded.source,
			updated_at = excluded.updated_at;
	`)
	if err != nil {
		return fmt.Errorf("prepare import stmt: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for _, h := range hashes {
		norm := strings.ToLower(strings.TrimSpace(h.SHA256))
		up := h.UpdatedAt
		if up.IsZero() {
			up = now
		}
		_, err = stmt.Exec(norm, h.Classification, h.ThreatName, h.Source, up)
		if err != nil {
			return fmt.Errorf("import hash %s: %w", h.SHA256, err)
		}
	}

	return tx.Commit()
}
