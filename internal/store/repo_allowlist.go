package store

import (
	"fmt"
	"time"
)

// AllowlistRecord represents a user-whitelisted package and certificate fingerprint.
type AllowlistRecord struct {
	ID         string    `json:"id"`
	Package    string    `json:"package"`
	CertSHA256 string    `json:"certSha256"`
	Reason     string    `json:"reason"`
	AddedAt    time.Time `json:"addedAt"`
}

// AddAllowlistEntry adds a package to the persistent allowlist.
func (s *Store) AddAllowlistEntry(entry AllowlistRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.AddedAt.IsZero() {
		entry.AddedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO allowlist (id, package_name, cert_sha256, reason, added_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			package_name = excluded.package_name,
			cert_sha256 = excluded.cert_sha256,
			reason = excluded.reason;
	`
	_, err := s.db.Exec(query, entry.ID, entry.Package, entry.CertSHA256, entry.Reason, entry.AddedAt)
	if err != nil {
		return fmt.Errorf("add allowlist entry %s: %w", entry.Package, err)
	}
	return nil
}

// RemoveAllowlistEntry deletes an entry by ID.
func (s *Store) RemoveAllowlistEntry(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM allowlist WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("remove allowlist entry %s: %w", id, err)
	}
	return nil
}

// ListAllowlist returns all active allowlist entries.
func (s *Store) ListAllowlist() ([]AllowlistRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := "SELECT id, package_name, cert_sha256, reason, added_at FROM allowlist ORDER BY added_at DESC"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list allowlist: %w", err)
	}
	defer rows.Close()

	var list []AllowlistRecord
	for rows.Next() {
		var entry AllowlistRecord
		err := rows.Scan(&entry.ID, &entry.Package, &entry.CertSHA256, &entry.Reason, &entry.AddedAt)
		if err != nil {
			return nil, fmt.Errorf("scan allowlist row: %w", err)
		}
		list = append(list, entry)
	}
	return list, rows.Err()
}

// IsAllowlisted checks if a package name and certificate match an allowlist entry.
func (s *Store) IsAllowlisted(pkg, certSHA string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT COUNT(*) FROM allowlist
		WHERE package_name = ? AND (cert_sha256 = '' OR cert_sha256 = ?)
	`
	var count int
	err := s.db.QueryRow(query, pkg, certSHA).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check allowlist for %s: %w", pkg, err)
	}
	return count > 0, nil
}
