package version_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/EnoughForYou/Jerion/internal/version"
)

type wailsConfig struct {
	Info struct {
		ProductVersion string `json:"productVersion"`
	} `json:"info"`
}

type packageJSON struct {
	Version string `json:"version"`
}

func TestVersionConsistency(t *testing.T) {
	if strings.Contains(version.AppVersion, "dev") {
		t.Skip("skipping version consistency check for dev build")
	}

	// 1. Check wails.json
	wailsPath := filepath.Join("..", "..", "wails.json")
	wailsData, err := os.ReadFile(wailsPath)
	if err != nil {
		t.Fatalf("failed to read wails.json: %v", err)
	}
	var wCfg wailsConfig
	if err := json.Unmarshal(wailsData, &wCfg); err != nil {
		t.Fatalf("failed to parse wails.json: %v", err)
	}
	if wCfg.Info.ProductVersion != version.AppVersion {
		t.Errorf("wails.json productVersion mismatch: got %q, want %q", wCfg.Info.ProductVersion, version.AppVersion)
	}

	// 2. Check frontend/package.json
	pkgPath := filepath.Join("..", "..", "frontend", "package.json")
	pkgData, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatalf("failed to read frontend/package.json: %v", err)
	}
	var pJSON packageJSON
	if err := json.Unmarshal(pkgData, &pJSON); err != nil {
		t.Fatalf("failed to parse frontend/package.json: %v", err)
	}
	if pJSON.Version != version.AppVersion {
		t.Errorf("frontend/package.json version mismatch: got %q, want %q", pJSON.Version, version.AppVersion)
	}

	// 3. Check CHANGELOG.md top released version
	changelogPath := filepath.Join("..", "..", "CHANGELOG.md")
	clFile, err := os.Open(changelogPath)
	if err != nil {
		t.Fatalf("failed to open CHANGELOG.md: %v", err)
	}
	defer clFile.Close()

	reVersion := regexp.MustCompile(`^## \[([0-9]+\.[0-9]+\.[0-9]+.*?)\]`)
	scanner := bufio.NewScanner(clFile)
	var latestChangelogVersion string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "## [Unreleased]") {
			continue
		}
		if matches := reVersion.FindStringSubmatch(line); len(matches) > 1 {
			latestChangelogVersion = matches[1]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("failed to scan CHANGELOG.md: %v", err)
	}

	if latestChangelogVersion == "" {
		t.Fatalf("no released version found in CHANGELOG.md")
	}
	if latestChangelogVersion != version.AppVersion {
		t.Errorf("CHANGELOG.md latest version mismatch: got %q, want %q", latestChangelogVersion, version.AppVersion)
	}
}
