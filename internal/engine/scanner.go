package engine

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/EnoughForYou/Jerion/internal/apk"
	"github.com/EnoughForYou/Jerion/internal/intel"
	"github.com/EnoughForYou/Jerion/internal/inventory"
	"github.com/EnoughForYou/Jerion/internal/logx"
	"github.com/EnoughForYou/Jerion/internal/rules"
	"github.com/EnoughForYou/Jerion/internal/store"
)

// Scanner coordinates end-to-end device threat analysis.
type Scanner struct {
	inventory *inventory.Service
	puller    *inventory.Puller
	inspector *apk.Inspector
	evaluator *Evaluator
	intel     *intel.Aggregator
	store     *store.Store
	workDir   string
	logger    *logx.Logger
}

// NewScanner constructs an operational threat Scanner.
func NewScanner(
	invSvc *inventory.Service,
	puller *inventory.Puller,
	inspector *apk.Inspector,
	intelAgg *intel.Aggregator,
	s *store.Store,
	workDir string,
	logger *logx.Logger,
) (*Scanner, error) {
	if logger == nil {
		logger = logx.GetDefault()
	}
	if inspector == nil {
		inspector = apk.NewInspector()
	}

	loadedRules, err := rules.LoadEmbeddedRules()
	if err != nil {
		return nil, fmt.Errorf("load rules for scanner: %w", err)
	}

	return &Scanner{
		inventory: invSvc,
		puller:    puller,
		inspector: inspector,
		evaluator: NewEvaluator(loadedRules),
		intel:     intelAgg,
		store:     s,
		workDir:   workDir,
		logger:    logger,
	}, nil
}

// Scan performs a Quick or Deep scan against a specified Android device serial.
func (s *Scanner) Scan(
	ctx context.Context,
	serial string,
	opts ScanOptions,
	onProgress func(ScanProgress),
) (*ScanResult, error) {
	startTime := time.Now().UTC()
	scanID := opts.ScanID
	if scanID == "" {
		scanID = fmt.Sprintf("scan-%d", startTime.UnixNano())
	}

	if onProgress != nil {
		onProgress(ScanProgress{Step: "Discovering installed applications", Progress: 10})
	}

	// 1. Fetch package list
	userOnly := !opts.IncludeSystem
	apps, err := s.inventory.ListPackages(ctx, serial, userOnly)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}

	totalApps := len(apps)
	if totalApps == 0 {
		return &ScanResult{
			ID:           scanID,
			DeviceSerial: serial,
			Mode:         opts.Mode,
			StartedAt:    startTime,
			FinishedAt:   time.Now().UTC(),
			RulesVersion: "1.0.0",
			AppVersion:   "0.4.0",
			Findings:     []Finding{},
			TotalApps:    0,
		}, nil
	}

	var allFindings []Finding
	isPartial := false

	// 2. Evaluate packages
	for idx, basicApp := range apps {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		pkgName := basicApp.Package
		percent := 10 + int(float64(idx+1)/float64(totalApps)*70) // 10% to 80%

		if onProgress != nil {
			stepName := "Inspecting package manifests"
			if opts.Mode == "deep" {
				stepName = "Extracting and analyzing APK binaries"
			}
			onProgress(ScanProgress{
				Step:           stepName,
				Progress:       percent,
				CurrentPackage: &pkgName,
			})
		}

		// Retrieve detailed package information
		detailApp, err := s.inventory.GetPackageDetail(ctx, serial, pkgName)
		if err != nil || detailApp == nil {
			detailApp = &basicApp
		}

		var apkDetail *apk.APKDetail

		// If Deep Scan, extract and inspect APK binary
		if opts.Mode == "deep" && s.puller != nil {
			pulledDir := filepath.Join(s.workDir, "pulled_apks")
			basePath, _, _, pullErr := s.puller.Pull(ctx, serial, pkgName, pulledDir)
			if pullErr == nil && basePath != "" {
				inspected, inspErr := s.inspector.Inspect(basePath)
				if inspErr == nil && inspected != nil {
					apkDetail = inspected

					// Check hash in threat intel
					if s.intel != nil && inspected.SHA256 != "" {
						intelRes, _ := s.intel.Lookup(ctx, inspected.SHA256)
						if intelRes != nil && intelRes.Classification != intel.ClassClean {
							allFindings = append(allFindings, Finding{
								ID:                fmt.Sprintf("%s-f-intel-%s", scanID, pkgName),
								RuleID:            "INTEL-HASH-001",
								Package:           pkgName,
								Category:          string(intelRes.Classification),
								Severity:          "critical",
								Confidence:        "high",
								Weight:            99,
								Title:             fmt.Sprintf("Known Threat Signature: %s", intelRes.ThreatName),
								WhyFlagged:        fmt.Sprintf("File SHA-256 matches threat intelligence signature from %s.", intelRes.Source),
								FalsePositiveHint: "Verify certificate signer against vendor release keys.",
								Evidence: []Evidence{
									{Source: "intel", Snippet: fmt.Sprintf("SHA-256: %s", inspected.SHA256)},
								},
								Actions: []string{"uninstall"},
							})
						}
					}
				} else {
					s.logger.Warn("deep scan: inspect apk failed", "package", pkgName, "err", inspErr)
					isPartial = true
				}
			} else {
				s.logger.Warn("deep scan: pull apk failed", "package", pkgName, "err", pullErr)
				isPartial = true
			}
		}

		// Evaluate declarative rules
		findings := s.evaluator.EvaluateApp(*detailApp, apkDetail)
		for _, f := range findings {
			f.ID = fmt.Sprintf("%s-%s", scanID, f.ID)
			allFindings = append(allFindings, f)
		}
	}

	if onProgress != nil {
		onProgress(ScanProgress{Step: "Consolidating security findings", Progress: 90})
	}

	finishTime := time.Now().UTC()
	result := &ScanResult{
		ID:           scanID,
		DeviceSerial: serial,
		Mode:         opts.Mode,
		StartedAt:    startTime,
		FinishedAt:   finishTime,
		RulesVersion: "1.0.0",
		AppVersion:   "0.4.0",
		Findings:     allFindings,
		Partial:      isPartial,
		TotalApps:    totalApps,
	}

	// 3. Persist scan result in database if store available
	if s.store != nil {
		var storeFindings []store.FindingRecord
		for _, f := range allFindings {
			var storeEvidence []store.EvidenceRecord
			for _, ev := range f.Evidence {
				storeEvidence = append(storeEvidence, store.EvidenceRecord{
					Source:  ev.Source,
					Snippet: ev.Snippet,
				})
			}
			storeFindings = append(storeFindings, store.FindingRecord{
				ID:                f.ID,
				ScanID:            scanID,
				RuleID:            f.RuleID,
				Package:           f.Package,
				Category:          f.Category,
				Severity:          f.Severity,
				Confidence:        f.Confidence,
				Weight:            f.Weight,
				Title:             f.Title,
				WhyFlagged:        f.WhyFlagged,
				FalsePositiveHint: f.FalsePositiveHint,
				Evidence:          storeEvidence,
				Actions:           f.Actions,
			})
		}

		err := s.store.SaveScanResult(store.ScanRecord{
			ID:           scanID,
			DeviceSerial: serial,
			Mode:         opts.Mode,
			StartedAt:    startTime,
			FinishedAt:   finishTime,
			RulesVersion: "1.0.0",
			AppVersion:   "0.4.0",
			Partial:      isPartial,
			TotalApps:    totalApps,
			Findings:     storeFindings,
		})
		if err != nil {
			s.logger.Error("engine: failed to save scan result to database", "scanId", scanID, "err", err)
			return result, fmt.Errorf("failed to save scan result to history: %w", err)
		}
	}

	if onProgress != nil {
		onProgress(ScanProgress{Step: "Scan completed", Progress: 100})
	}

	return result, nil
}
