package intel

import (
	"context"
	"errors"
	"strings"

	"github.com/EnoughForYou/Jerion/internal/secrets"
	"github.com/EnoughForYou/Jerion/internal/store"
)

var ErrHashNotFound = errors.New("hash not found in threat intelligence")

// ThreatClassification represents the verdict on a file hash.
type ThreatClassification string

const (
	ClassMalware ThreatClassification = "malware"
	ClassAdware  ThreatClassification = "adware"
	ClassPUP     ThreatClassification = "pup"
	ClassClean   ThreatClassification = "clean"
	ClassUnknown ThreatClassification = "unknown"
)

// IntelResult holds threat intelligence for a file hash.
type IntelResult struct {
	SHA256         string               `json:"sha256"`
	Classification ThreatClassification `json:"classification"`
	ThreatName     string               `json:"threatName,omitempty"`
	Source         string               `json:"source"`
	Positives      int                  `json:"positives,omitempty"`
	Total          int                  `json:"total,omitempty"`
}

// Provider queries threat intelligence.
type Provider interface {
	Lookup(ctx context.Context, sha256Hex string) (*IntelResult, error)
}

// LocalProvider queries local SQLite intel database.
type LocalProvider struct {
	store *store.Store
}

// NewLocalProvider creates a local SQLite threat intel provider.
func NewLocalProvider(s *store.Store) *LocalProvider {
	return &LocalProvider{store: s}
}

func (l *LocalProvider) Lookup(ctx context.Context, sha256Hex string) (*IntelResult, error) {
	if l.store == nil {
		return nil, ErrHashNotFound
	}
	record, err := l.store.LookupIntelHash(sha256Hex)
	if err != nil {
		return nil, ErrHashNotFound
	}

	return &IntelResult{
		SHA256:         record.SHA256,
		Classification: ThreatClassification(record.Classification),
		ThreatName:     record.ThreatName,
		Source:         record.Source,
	}, nil
}

// Aggregator coordinates local and remote intelligence providers.
type Aggregator struct {
	local   *LocalProvider
	secrets *secrets.Manager
	clamav  *ClamAVClient
}

// NewAggregator creates an intelligence aggregator.
func NewAggregator(s *store.Store, sm *secrets.Manager) *Aggregator {
	return &Aggregator{
		local:   NewLocalProvider(s),
		secrets: sm,
	}
}

// SetClamAVClient configures the ClamAV daemon client.
func (a *Aggregator) SetClamAVClient(c *ClamAVClient) {
	a.clamav = c
}

// GetClamAVClient returns the configured ClamAV daemon client, if any.
func (a *Aggregator) GetClamAVClient() *ClamAVClient {
	return a.clamav
}

// ScanFileWithClamAV scans an APK file using ClamAV if configured.
func (a *Aggregator) ScanFileWithClamAV(ctx context.Context, filePath string) (*IntelResult, error) {
	if a.clamav == nil {
		return nil, ErrClamAVUnavailable
	}
	res, err := a.clamav.ScanFile(ctx, filePath)
	if err != nil {
		return nil, err
	}
	if res.Infected {
		return &IntelResult{
			Classification: ClassMalware,
			ThreatName:     res.Virus,
			Source:         "clamav",
			Positives:      1,
			Total:          1,
		}, nil
	}
	return &IntelResult{
		Classification: ClassClean,
		Source:         "clamav",
	}, nil
}

// Lookup queries local threat intel cache first.
func (a *Aggregator) Lookup(ctx context.Context, sha256Hex string) (*IntelResult, error) {
	norm := strings.ToLower(strings.TrimSpace(sha256Hex))
	if a.local != nil {
		res, err := a.local.Lookup(ctx, norm)
		if err == nil && res != nil {
			return res, nil
		}
	}
	return nil, ErrHashNotFound
}
