package logx

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents the severity level of a log message.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a level string into a Level.
func ParseLevel(lvl string) Level {
	switch strings.ToUpper(strings.TrimSpace(lvl)) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger provides structured, thread-safe logging.
type Logger struct {
	mu         sync.Mutex
	out        io.Writer
	minLevel   Level
	fields     map[string]any
	showCaller bool
}

var (
	defaultLoggerMu sync.RWMutex
	defaultLogger   *Logger
)

func init() {
	defaultLogger = New(os.Stdout, LevelInfo)
}

// SetDefault sets the package-level default logger.
func SetDefault(l *Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	defaultLogger = l
}

// GetDefault returns the package-level default logger.
func GetDefault() *Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

// New creates a new Logger writing to out with minLevel.
func New(out io.Writer, minLevel Level) *Logger {
	if out == nil {
		out = io.Discard
	}
	return &Logger{
		out:        out,
		minLevel:   minLevel,
		fields:     make(map[string]any),
		showCaller: true,
	}
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(lvl Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = lvl
}

// Level returns the current minimum log level.
func (l *Logger) Level() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.minLevel
}

// With returns a child logger with additional persistent fields.
func (l *Logger) With(keyValues ...any) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	childFields := make(map[string]any, len(l.fields)+len(keyValues)/2)
	for k, v := range l.fields {
		childFields[k] = v
	}

	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			k := fmt.Sprintf("%v", keyValues[i])
			childFields[k] = keyValues[i+1]
		}
	}

	return &Logger{
		out:        l.out,
		minLevel:   l.minLevel,
		fields:     childFields,
		showCaller: l.showCaller,
	}
}

func (l *Logger) log(level Level, msg string, keyValues ...any) {
	l.mu.Lock()
	if level < l.minLevel {
		l.mu.Unlock()
		return
	}
	out := l.out
	showCaller := l.showCaller

	// Merge persistent fields and call-site fields
	allFields := make([]string, 0, len(l.fields)+len(keyValues)/2)
	for k, v := range l.fields {
		allFields = append(allFields, fmt.Sprintf("%s=%v", k, v))
	}
	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			allFields = append(allFields, fmt.Sprintf("%v=%v", keyValues[i], keyValues[i+1]))
		}
	}
	l.mu.Unlock()

	now := time.Now().Format("2006-01-02 15:04:05.000")
	var callerStr string
	if showCaller {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			callerStr = fmt.Sprintf(" %s:%d", filepath.Base(file), line)
		}
	}

	fieldsPart := ""
	if len(allFields) > 0 {
		fieldsPart = " [" + strings.Join(allFields, " ") + "]"
	}

	line := fmt.Sprintf("%s [%s]%s %s%s\n", now, level.String(), callerStr, msg, fieldsPart)
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = io.WriteString(out, line)
}

// Debug logs at LevelDebug.
func (l *Logger) Debug(msg string, keyValues ...any) {
	l.log(LevelDebug, msg, keyValues...)
}

// Info logs at LevelInfo.
func (l *Logger) Info(msg string, keyValues ...any) {
	l.log(LevelInfo, msg, keyValues...)
}

// Warn logs at LevelWarn.
func (l *Logger) Warn(msg string, keyValues ...any) {
	l.log(LevelWarn, msg, keyValues...)
}

// Error logs at LevelError.
func (l *Logger) Error(msg string, keyValues ...any) {
	l.log(LevelError, msg, keyValues...)
}

// DebugADB logs an executed raw ADB command for auditing/troubleshooting per NFR-09.
func (l *Logger) DebugADB(serial, cmd string, args []string, duration time.Duration, exitCode int, err error) {
	errStr := "nil"
	if err != nil {
		errStr = err.Error()
	}
	l.Debug("ADB_EXEC",
		"serial", serial,
		"cmd", fmt.Sprintf("%s %s", cmd, strings.Join(args, " ")),
		"duration", duration.String(),
		"exit_code", exitCode,
		"err", errStr,
	)
}

// Global package-level shortcuts delegating to GetDefault()
func Debug(msg string, keyValues ...any) { GetDefault().log(LevelDebug, msg, keyValues...) }
func Info(msg string, keyValues ...any)  { GetDefault().log(LevelInfo, msg, keyValues...) }
func Warn(msg string, keyValues ...any)  { GetDefault().log(LevelWarn, msg, keyValues...) }
func Error(msg string, keyValues ...any) { GetDefault().log(LevelError, msg, keyValues...) }

// RotatingFileWriter provides simple size-bounded local file logging with rotation.
type RotatingFileWriter struct {
	mu       sync.Mutex
	dir      string
	baseName string
	maxBytes int64
	file     *os.File
	curSize  int64
}

// NewRotatingFileWriter opens or creates a rotating log file in dir.
func NewRotatingFileWriter(dir, baseName string, maxBytes int64) (*RotatingFileWriter, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	if maxBytes <= 0 {
		maxBytes = 10 * 1024 * 1024 // 10MB default
	}

	rw := &RotatingFileWriter{
		dir:      dir,
		baseName: baseName,
		maxBytes: maxBytes,
	}

	if err := rw.openCurrent(); err != nil {
		return nil, err
	}
	return rw, nil
}

func (rw *RotatingFileWriter) openCurrent() error {
	filePath := filepath.Join(rw.dir, rw.baseName+".log")
	fi, err := os.Stat(filePath)
	if err == nil {
		rw.curSize = fi.Size()
	} else if os.IsNotExist(err) {
		rw.curSize = 0
	} else {
		return err
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	rw.file = f
	return nil
}

func (rw *RotatingFileWriter) Write(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.file == nil {
		if err := rw.openCurrent(); err != nil {
			return 0, err
		}
	}

	if rw.curSize+int64(len(p)) > rw.maxBytes {
		_ = rw.file.Close()
		oldPath := filepath.Join(rw.dir, rw.baseName+".log")
		backupPath := filepath.Join(rw.dir, fmt.Sprintf("%s-%s.log", rw.baseName, time.Now().Format("20060102-150405")))
		_ = os.Rename(oldPath, backupPath)
		if err := rw.openCurrent(); err != nil {
			return 0, err
		}
	}

	n, err = rw.file.Write(p)
	rw.curSize += int64(n)
	return n, err
}

// Close closes the underlying log file.
func (rw *RotatingFileWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if rw.file != nil {
		err := rw.file.Close()
		rw.file = nil
		return err
	}
	return nil
}

// ContextWithLogger stores logger in context.
type ctxKey struct{}

func WithContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*Logger); ok && l != nil {
		return l
	}
	return GetDefault()
}
