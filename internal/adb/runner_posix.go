//go:build !windows

package adb

import "os/exec"

func setHideWindow(cmd *exec.Cmd) {
	// No-op on non-Windows platforms
}
