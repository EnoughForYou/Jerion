package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

// Store encapsulates the SQLite database connection and operations.
type Store struct {
	mu sync.RWMutex
	db *sql.DB
}

// DefaultDBPath returns the standard persistent database path per platform.
func DefaultDBPath() (string, error) {
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "jerion.db", nil
		}
		appData = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(appData, "Jerion")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("create jerion app data directory: %w", err)
	}
	return filepath.Join(dir, "jerion.db"), nil
}

// Open initializes or connects to a SQLite database file with WAL mode.
func Open(dataSourceName string) (*Store, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// Restrict connection pool to maintain SQLite single-writer safety
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite db: %w", err)
	}

	// Enforce busy timeout and foreign keys
	if _, err := db.Exec("PRAGMA busy_timeout = 5000; PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set sqlite pragmas: %w", err)
	}

	// For disk files, activate WAL mode
	if dataSourceName != ":memory:" && !strings.Contains(dataSourceName, "memory") {
		_, _ = db.Exec("PRAGMA journal_mode = WAL;")
	}

	s := &Store{db: db}
	if err := s.RunMigrations(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run sqlite migrations: %w", err)
	}

	return s, nil
}

// OpenMemory creates an in-memory SQLite store for unit testing.
func OpenMemory() (*Store, error) {
	return Open(":memory:")
}

// DB returns the underlying *sql.DB.
func (s *Store) DB() *sql.DB {
	return s.db
}

// Close closes the database connection.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		return err
	}
	return nil
}
