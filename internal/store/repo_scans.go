package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrScanNotFound = errors.New("scan result not found")

// EvidenceRecord represents a technical proof point.
type EvidenceRecord struct {
	Source  string `json:"source"`
	Snippet string `json:"snippet"`
}

// FindingRecord captures an individual security or adware finding.
type FindingRecord struct {
	ID                string           `json:"id"`
	ScanID            string           `json:"scanId"`
	RuleID            string           `json:"ruleId"`
	Package           string           `json:"package"`
	Category          string           `json:"category"`
	Severity          string           `json:"severity"`
	Confidence        string           `json:"confidence"`
	Weight            int              `json:"weight"`
	Title             string           `json:"title"`
	WhyFlagged        string           `json:"whyFlagged"`
	FalsePositiveHint string           `json:"falsePositiveHint"`
	Evidence          []EvidenceRecord `json:"evidence"`
	Actions           []string         `json:"actions"`
}

// ScanRecord represents an executed scan result.
type ScanRecord struct {
	ID           string          `json:"id"`
	DeviceSerial string          `json:"deviceSerial"`
	Mode         string          `json:"mode"`
	StartedAt    time.Time       `json:"startedAt"`
	FinishedAt   time.Time       `json:"finishedAt"`
	RulesVersion string          `json:"rulesVersion"`
	AppVersion   string          `json:"appVersion"`
	Partial      bool            `json:"partial"`
	Findings     []FindingRecord `json:"findings"`
}

// ScanSummaryRecord provides lightweight metadata for scan history lists.
type ScanSummaryRecord struct {
	ID                string    `json:"id"`
	DeviceSerial      string    `json:"deviceSerial"`
	Mode              string    `json:"mode"`
	StartedAt         time.Time `json:"startedAt"`
	FinishedAt        time.Time `json:"finishedAt"`
	FindingCount      int       `json:"findingCount"`
	HighSeverityCount int       `json:"highSeverityCount"`
}

// SaveScanResult stores a complete scan and its associated findings within a transaction.
func (s *Store) SaveScanResult(scan ScanRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin scan tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	partialInt := 0
	if scan.Partial {
		partialInt = 1
	}

	insertScan := `
		INSERT INTO scan_results (id, device_serial, mode, started_at, finished_at, rules_version, app_version, partial)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			finished_at = excluded.finished_at,
			partial = excluded.partial;
	`
	_, err = tx.Exec(insertScan,
		scan.ID, scan.DeviceSerial, scan.Mode, scan.StartedAt, scan.FinishedAt, scan.RulesVersion, scan.AppVersion, partialInt,
	)
	if err != nil {
		return fmt.Errorf("insert scan_result: %w", err)
	}

	// Delete existing findings for scan if updating
	_, err = tx.Exec("DELETE FROM scan_findings WHERE scan_id = ?", scan.ID)
	if err != nil {
		return fmt.Errorf("clear existing findings: %w", err)
	}

	insertFinding := `
		INSERT INTO scan_findings (
			id, scan_id, rule_id, package_name, category, severity, confidence, weight,
			title, why_flagged, fp_hint, evidence_json, actions_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	stmt, err := tx.Prepare(insertFinding)
	if err != nil {
		return fmt.Errorf("prepare finding insert: %w", err)
	}
	defer stmt.Close()

	for _, f := range scan.Findings {
		evJSON, err := json.Marshal(f.Evidence)
		if err != nil {
			return fmt.Errorf("marshal evidence for finding %s: %w", f.ID, err)
		}
		actJSON, err := json.Marshal(f.Actions)
		if err != nil {
			return fmt.Errorf("marshal actions for finding %s: %w", f.ID, err)
		}

		_, err = stmt.Exec(
			f.ID, scan.ID, f.RuleID, f.Package, f.Category, f.Severity, f.Confidence, f.Weight,
			f.Title, f.WhyFlagged, f.FalsePositiveHint, string(evJSON), string(actJSON),
		)
		if err != nil {
			return fmt.Errorf("insert finding %s: %w", f.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit scan transaction: %w", err)
	}
	return nil
}

// GetScanResult retrieves a scan and all its child findings.
func (s *Store) GetScanResult(id string) (*ScanRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var scan ScanRecord
	var partialInt int
	query := `
		SELECT id, device_serial, mode, started_at, finished_at, rules_version, app_version, partial
		FROM scan_results WHERE id = ?
	`
	err := s.db.QueryRow(query, id).Scan(
		&scan.ID, &scan.DeviceSerial, &scan.Mode, &scan.StartedAt, &scan.FinishedAt, &scan.RulesVersion, &scan.AppVersion, &partialInt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrScanNotFound
		}
		return nil, fmt.Errorf("get scan %s: %w", id, err)
	}
	scan.Partial = partialInt == 1

	findingsQuery := `
		SELECT id, scan_id, rule_id, package_name, category, severity, confidence, weight,
		       title, why_flagged, fp_hint, evidence_json, actions_json
		FROM scan_findings WHERE scan_id = ?
		ORDER BY weight DESC, severity ASC
	`
	rows, err := s.db.Query(findingsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("query findings for scan %s: %w", id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var f FindingRecord
		var evJSON, actJSON string
		err := rows.Scan(
			&f.ID, &f.ScanID, &f.RuleID, &f.Package, &f.Category, &f.Severity, &f.Confidence, &f.Weight,
			&f.Title, &f.WhyFlagged, &f.FalsePositiveHint, &evJSON, &actJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan finding row: %w", err)
		}

		if len(evJSON) > 0 {
			_ = json.Unmarshal([]byte(evJSON), &f.Evidence)
		}
		if len(actJSON) > 0 {
			_ = json.Unmarshal([]byte(actJSON), &f.Actions)
		}

		scan.Findings = append(scan.Findings, f)
	}

	return &scan, rows.Err()
}

// ListScanSummaries returns summary records for all completed scans of a device (or all devices if empty).
func (s *Store) ListScanSummaries(deviceSerial string) ([]ScanSummaryRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT s.id, s.device_serial, s.mode, s.started_at, s.finished_at,
		       COUNT(f.id) AS finding_count,
		       COALESCE(SUM(CASE WHEN f.severity IN ('high', 'critical') THEN 1 ELSE 0 END), 0) AS high_count
		FROM scan_results s
		LEFT JOIN scan_findings f ON s.id = f.scan_id
	`
	var args []any
	if len(deviceSerial) > 0 {
		query += " WHERE s.device_serial = ? "
		args = append(args, deviceSerial)
	}
	query += " GROUP BY s.id ORDER BY s.started_at DESC "

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list scan summaries: %w", err)
	}
	defer rows.Close()

	var summaries []ScanSummaryRecord
	for rows.Next() {
		var sum ScanSummaryRecord
		err := rows.Scan(
			&sum.ID, &sum.DeviceSerial, &sum.Mode, &sum.StartedAt, &sum.FinishedAt,
			&sum.FindingCount, &sum.HighSeverityCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan summary row: %w", err)
		}
		summaries = append(summaries, sum)
	}
	return summaries, rows.Err()
}

// DeleteScanResult removes a scan result and all cascading findings.
func (s *Store) DeleteScanResult(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM scan_results WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete scan %s: %w", id, err)
	}
	return nil
}
