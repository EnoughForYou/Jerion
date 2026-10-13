package secrets

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the default keyring service identifier for Jerion.
	ServiceName = "Jerion"
	// KeyVirusTotalAPI is the key user identifier for VirusTotal API token.
	KeyVirusTotalAPI = "virustotal_api_key"
	// ProbeUser is a temporary key used during P6 Preflight check.
	ProbeUser = "jerion_keychain_probe"
)

var (
	ErrNotFound      = errors.New("secret not found in keychain")
	ErrNotSupported   = errors.New("keychain service not available on this system")
	ErrInvalidSecret = errors.New("secret value cannot be empty")
)

// Keyring defines the interface for secure credential storage.
type Keyring interface {
	Set(service, user, password string) error
	Get(service, user string) (string, error)
	Delete(service, user string) error
}

// OSKeyring interacts with the operating system's native credential store.
type OSKeyring struct{}

// NewOSKeyring creates an OS-backed Keyring.
func NewOSKeyring() *OSKeyring {
	return &OSKeyring{}
}

func (k *OSKeyring) Set(service, user, password string) error {
	if password == "" {
		return ErrInvalidSecret
	}
	if err := keyring.Set(service, user, password); err != nil {
		return fmt.Errorf("keyring set: %w", err)
	}
	return nil
}

func (k *OSKeyring) Get(service, user string) (string, error) {
	val, err := keyring.Get(service, user)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("keyring get: %w", err)
	}
	return val, nil
}

func (k *OSKeyring) Delete(service, user string) error {
	err := keyring.Delete(service, user)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("keyring delete: %w", err)
	}
	return nil
}

// MemoryKeyring is an in-memory Keyring implementation used for tests or environments without native keychain.
type MemoryKeyring struct {
	mu    sync.RWMutex
	store map[string]string
}

// NewMemoryKeyring returns a new in-memory Keyring.
func NewMemoryKeyring() *MemoryKeyring {
	return &MemoryKeyring{
		store: make(map[string]string),
	}
}

func (m *MemoryKeyring) key(service, user string) string {
	return fmt.Sprintf("%s:%s", service, user)
}

func (m *MemoryKeyring) Set(service, user, password string) error {
	if password == "" {
		return ErrInvalidSecret
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[m.key(service, user)] = password
	return nil
}

func (m *MemoryKeyring) Get(service, user string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.store[m.key(service, user)]
	if !ok {
		return "", ErrNotFound
	}
	return val, nil
}

func (m *MemoryKeyring) Delete(service, user string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, m.key(service, user))
	return nil
}

// Manager wraps Keyring operations specific to Jerion domains.
type Manager struct {
	kr Keyring
}

// NewManager creates a Manager using the specified Keyring.
func NewManager(kr Keyring) *Manager {
	if kr == nil {
		kr = NewOSKeyring()
	}
	return &Manager{kr: kr}
}

// ProbeTest executes a round-trip write, read, and delete to verify keychain health (Preflight P6).
func (m *Manager) ProbeTest() error {
	probeValue := "test_probe_value"
	if err := m.kr.Set(ServiceName, ProbeUser, probeValue); err != nil {
		return fmt.Errorf("keychain write probe failed: %w", err)
	}

	val, err := m.kr.Get(ServiceName, ProbeUser)
	if err != nil {
		_ = m.kr.Delete(ServiceName, ProbeUser)
		return fmt.Errorf("keychain read probe failed: %w", err)
	}
	if val != probeValue {
		_ = m.kr.Delete(ServiceName, ProbeUser)
		return errors.New("keychain probe value mismatch")
	}

	if err := m.kr.Delete(ServiceName, ProbeUser); err != nil {
		return fmt.Errorf("keychain delete probe failed: %w", err)
	}
	return nil
}

// SaveVirusTotalKey stores the VirusTotal API key.
func (m *Manager) SaveVirusTotalKey(key string) error {
	return m.kr.Set(ServiceName, KeyVirusTotalAPI, key)
}

// GetVirusTotalKey retrieves the VirusTotal API key.
func (m *Manager) GetVirusTotalKey() (string, error) {
	return m.kr.Get(ServiceName, KeyVirusTotalAPI)
}

// DeleteVirusTotalKey removes the VirusTotal API key.
func (m *Manager) DeleteVirusTotalKey() error {
	return m.kr.Delete(ServiceName, KeyVirusTotalAPI)
}

// HasVirusTotalKey returns whether a VirusTotal API key is configured.
func (m *Manager) HasVirusTotalKey() bool {
	val, err := m.GetVirusTotalKey()
	return err == nil && len(val) > 0
}
