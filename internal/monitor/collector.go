package monitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/EnoughForYou/Jerion/internal/adb"
)

func (m *Manager) runCollectorLoop(ctx context.Context, serial string) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	var lastPackage string
	var lastOverlayPkg string

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stdout, _, err := m.runner.Run(ctx, serial, "shell", "dumpsys", "window")
			if err != nil {
				continue
			}

			outStr := string(stdout)
			focus := adb.ParseWindowFocus(outStr)

			now := time.Now().UTC().Format(time.RFC3339)

			// 1. Focus change event
			if focus.Package != "" && focus.Package != lastPackage {
				lastPackage = focus.Package
				evt := MonitorEvent{
					ID:        fmt.Sprintf("evt-focus-%d", time.Now().UnixNano()),
					Timestamp: now,
					Type:      EventForegroundSwitch,
					Package:   focus.Package,
					Details: map[string]any{
						"activity": focus.Activity,
						"rawFocus": focus.RawFocus,
					},
				}
				m.RecordEvent(evt)
			}

			// 2. Check for overlays
			if strings.Contains(outStr, "TYPE_APPLICATION_OVERLAY") {
				// Identify overlay owner if present in dumpsys window
				lines := strings.Split(outStr, "\n")
				foundOverlay := false
				for _, l := range lines {
					if strings.Contains(l, "TYPE_APPLICATION_OVERLAY") {
						trimmed := strings.TrimSpace(l)
						parts := strings.Fields(trimmed)
						for _, p := range parts {
							if strings.Contains(p, "/") {
								pkgCandidate := strings.Split(p, "/")[0]
								if pkgCandidate != "" && !stringsHasSystemPrefix(pkgCandidate) {
									foundOverlay = true
									if pkgCandidate != lastOverlayPkg {
										lastOverlayPkg = pkgCandidate
										evt := MonitorEvent{
											ID:        fmt.Sprintf("evt-overlay-%d", time.Now().UnixNano()),
											Timestamp: now,
											Type:      EventOverlayShown,
											Package:   pkgCandidate,
											Details: map[string]any{
												"windowType": "TYPE_APPLICATION_OVERLAY",
												"rawLine":    trimmed,
											},
										}
										m.RecordEvent(evt)
									}
									break
								}
							}
						}
						if foundOverlay {
							break
						}
					}
				}
				if !foundOverlay {
					lastOverlayPkg = ""
				}
			} else {
				lastOverlayPkg = ""
			}
		}
	}
}
