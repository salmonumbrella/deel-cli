package secrets

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/deel-cli/internal/config"
)

const (
	// CredentialRotationThreshold is the age after which credentials should be rotated
	CredentialRotationThreshold = 90 * 24 * time.Hour
)

var warnedAccounts sync.Map

// Store defines the interface for credential storage
type Store interface {
	Keys() ([]string, error)
	Set(name string, creds Credentials) error
	Get(name string) (Credentials, error)
	Delete(name string) error
	List() ([]Credentials, error)
}

// KeyringStore implements Store using the OS keychain
type KeyringStore struct {
	ring keyring.Keyring
}

// Credentials holds the authentication credentials for a Deel account
type Credentials struct {
	Name      string    `json:"name"`
	Token     string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type storedCredentials struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

// OpenDefault opens the default keyring store
func OpenDefault() (Store, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: config.AppName,
	})
	if err != nil {
		return nil, err
	}
	return &KeyringStore{ring: ring}, nil
}

// Keys returns all keys in the keyring
func (s *KeyringStore) Keys() ([]string, error) {
	return s.ring.Keys()
}

// Set stores credentials for an account
func (s *KeyringStore) Set(name string, creds Credentials) error {
	name = normalize(name)
	if name == "" {
		return fmt.Errorf("missing account name")
	}
	if creds.Token == "" {
		return fmt.Errorf("missing token")
	}
	if creds.CreatedAt.IsZero() {
		creds.CreatedAt = time.Now().UTC()
	}

	payload, err := json.Marshal(storedCredentials{
		Token:     creds.Token,
		CreatedAt: creds.CreatedAt,
	})
	if err != nil {
		return err
	}

	return s.ring.Set(keyring.Item{
		Key:  credentialKey(name),
		Data: payload,
	})
}

// Get retrieves credentials for an account
func (s *KeyringStore) Get(name string) (Credentials, error) {
	name = normalize(name)
	if name == "" {
		return Credentials{}, fmt.Errorf("missing account name")
	}
	item, err := s.ring.Get(credentialKey(name))
	if err != nil {
		return Credentials{}, err
	}
	var stored storedCredentials
	if err := json.Unmarshal(item.Data, &stored); err != nil {
		return Credentials{}, err
	}

	creds := Credentials{
		Name:      name,
		Token:     stored.Token,
		CreatedAt: stored.CreatedAt,
	}

	// Warn if credentials are older than 90 days (backwards compatible with zero time)
	// Only warn once per session per account to avoid spam
	if !creds.CreatedAt.IsZero() && time.Since(creds.CreatedAt) > CredentialRotationThreshold {
		if _, warned := warnedAccounts.LoadOrStore(name, true); !warned {
			slog.Warn("credentials over 90 days old, consider rotating", "account", name, "age_days", int(time.Since(creds.CreatedAt).Hours()/24))
		}
	}

	return creds, nil
}

// Delete removes credentials for an account
func (s *KeyringStore) Delete(name string) error {
	name = normalize(name)
	if name == "" {
		return fmt.Errorf("missing account name")
	}
	return s.ring.Remove(credentialKey(name))
}

// List returns all stored credentials
func (s *KeyringStore) List() ([]Credentials, error) {
	keys, err := s.Keys()
	if err != nil {
		return nil, err
	}
	var out []Credentials
	for _, k := range keys {
		name, ok := ParseCredentialKey(k)
		if !ok {
			continue
		}
		creds, err := s.Get(name)
		if err != nil {
			return nil, err
		}
		out = append(out, creds)
	}
	return out, nil
}

// ParseCredentialKey extracts the account name from a credential key
func ParseCredentialKey(k string) (name string, ok bool) {
	const prefix = "account:"
	if !strings.HasPrefix(k, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(k, prefix)
	if strings.TrimSpace(rest) == "" {
		return "", false
	}
	return rest, true
}

func credentialKey(name string) string {
	return fmt.Sprintf("account:%s", name)
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
