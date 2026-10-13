package logx

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	logger.Debug("should not appear")
	if buf.Len() > 0 {
		t.Fatalf("expected debug to be ignored, got: %s", buf.String())
	}

	logger.Info("info message", "key", "val")
	if !strings.Contains(buf.String(), "[INFO]") || !strings.Contains(buf.String(), "info message") || !strings.Contains(buf.String(), "key=val") {
		t.Fatalf("unexpected info log: %s", buf.String())
	}

	buf.Reset()
	logger.Warn("warning alert")
	if !strings.Contains(buf.String(), "[WARN]") || !strings.Contains(buf.String(), "warning alert") {
		t.Fatalf("unexpected warn log: %s", buf.String())
	}

	buf.Reset()
	logger.Error("error occurred")
	if !strings.Contains(buf.String(), "[ERROR]") || !strings.Contains(buf.String(), "error occurred") {
		t.Fatalf("unexpected error log: %s", buf.String())
	}
}

func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	base := New(&buf, LevelDebug)
	child := base.With("component", "adb_runner", "serial", "TEST-123")

	child.Debug("command executed", "exit", 0)
	out := buf.String()
	if !strings.Contains(out, "component=adb_runner") || !strings.Contains(out, "serial=TEST-123") || !strings.Contains(out, "exit=0") {
		t.Fatalf("missing fields in child logger: %s", out)
	}
}

func TestDebugADB(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelDebug)

	logger.DebugADB("DEVICE_001", "adb", []string{"shell", "getprop"}, 25*time.Millisecond, 0, nil)
	out := buf.String()
	if !strings.Contains(out, "ADB_EXEC") || !strings.Contains(out, "DEVICE_001") || !strings.Contains(out, "adb shell getprop") {
		t.Fatalf("unexpected DebugADB output: %s", out)
	}

	buf.Reset()
	logger.DebugADB("DEVICE_001", "adb", []string{"shell", "dumpsys"}, 15*time.Millisecond, 1, errors.New("timeout"))
	out = buf.String()
	if !strings.Contains(out, "err=timeout") || !strings.Contains(out, "exit_code=1") {
		t.Fatalf("unexpected DebugADB error output: %s", out)
	}
}

func TestRotatingFileWriter(t *testing.T) {
	tmpDir := t.TempDir()
	maxBytes := int64(200)

	rfw, err := NewRotatingFileWriter(tmpDir, "jerion_test", maxBytes)
	if err != nil {
		t.Fatalf("NewRotatingFileWriter failed: %v", err)
	}
	defer rfw.Close()

	// Write more than maxBytes to trigger rotation
	payload := strings.Repeat("A", 120) + "\n"
	if _, err := rfw.Write([]byte(payload)); err != nil {
		t.Fatalf("write 1 failed: %v", err)
	}
	if _, err := rfw.Write([]byte(payload)); err != nil {
		t.Fatalf("write 2 failed: %v", err)
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read tmpDir failed: %v", err)
	}

	// Should have at least the active .log and one rotated backup .log
	if len(files) < 2 {
		t.Fatalf("expected at least 2 log files after rotation, got %d", len(files))
	}
}

func TestContextLogger(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, LevelDebug)
	ctx := WithContext(context.Background(), l)

	extracted := FromContext(ctx)
	if extracted != l {
		t.Fatalf("expected extracted logger to match context logger")
	}

	defaultL := FromContext(context.Background())
	if defaultL != GetDefault() {
		t.Fatalf("expected default logger fallback")
	}
}

func TestParseLevel(t *testing.T) {
	if ParseLevel("DEBUG") != LevelDebug {
		t.Errorf("expected LevelDebug")
	}
	if ParseLevel("warn") != LevelWarn {
		t.Errorf("expected LevelWarn")
	}
	if ParseLevel("error") != LevelError {
		t.Errorf("expected LevelError")
	}
	if ParseLevel("unknown") != LevelInfo {
		t.Errorf("expected LevelInfo default")
	}
}
