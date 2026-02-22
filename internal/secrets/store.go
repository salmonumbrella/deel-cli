package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/99designs/keyring"
	"golang.org/x/term"

	"github.com/salmonumbrella/deel-cli/internal/config"
)

const (
	// CredentialRotationThreshold is the age after which credentials should be rotated
	CredentialRotationThreshold = 90 * 24 * time.Hour

	keyringPasswordEnv = "DEEL_KEYRING_PASSWORD"
)

var (
	warnedAccounts sync.Map
	errNoTTY       = errors.New("no TTY available for keyring file backend password prompt")
)

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
	ring, err := openKeyring(runtime.GOOS, os.Getenv("DBUS_SESSION_BUS_ADDRESS"))
	if err != nil {
		return nil, err
	}
	return &KeyringStore{ring: ring}, nil
}

func openKeyring(goos string, dbusAddr string) (keyring.Keyring, error) {
	keyringDir, err := ensureKeyringDir()
	if err != nil {
		return nil, fmt.Errorf("ensure keyring dir: %w", err)
	}

	var allowedBackends []keyring.BackendType
	if shouldForceFileBackend(goos, dbusAddr) {
		allowedBackends = []keyring.BackendType{keyring.FileBackend}
	}

	ring, err := keyring.Open(keyring.Config{
		ServiceName:      config.AppName,
		AllowedBackends:  allowedBackends,
		FileDir:          keyringDir,
		FilePasswordFunc: fileKeyringPasswordFunc(),
	})
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	return ring, nil
}

func ensureKeyringDir() (string, error) {
	keyringDir, err := resolveKeyringDir()
	if err != nil {
		return "", err
	}

	info, err := os.Stat(keyringDir)
	switch {
	case err == nil:
		if !info.IsDir() {
			return "", fmt.Errorf("keyring path %q is not a directory", keyringDir)
		}
	case os.IsNotExist(err):
		if err := os.MkdirAll(keyringDir, 0o700); err != nil {
			return "", fmt.Errorf("create keyring directory: %w", err)
		}
	case err != nil:
		return "", fmt.Errorf("inspect keyring directory: %w", err)
	}

	return keyringDir, nil
}

func resolveKeyringDir() (string, error) {
	if dir, ok, err := credentialsDirFromEnv(); err != nil {
		return "", err
	} else if ok {
		return dir, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate config directory: %w", err)
	}

	return filepath.Join(configDir, config.AppName, "keyring"), nil
}

func credentialsDirFromEnv() (dir string, ok bool, err error) {
	if explicit := strings.TrimSpace(os.Getenv(config.EnvCredentialsDir)); explicit != "" {
		expanded, err := keyring.ExpandTilde(explicit)
		if err != nil {
			return "", false, fmt.Errorf("expand %s: %w", config.EnvCredentialsDir, err)
		}
		return filepath.Clean(expanded), true, nil
	}

	if shared := strings.TrimSpace(os.Getenv(config.EnvOpenClawCredentialsDir)); shared != "" {
		expanded, err := keyring.ExpandTilde(shared)
		if err != nil {
			return "", false, fmt.Errorf("expand %s: %w", config.EnvOpenClawCredentialsDir, err)
		}
		return filepath.Join(filepath.Clean(expanded), config.AppName, "keyring"), true, nil
	}

	return "", false, nil
}

func shouldForceFileBackend(goos string, dbusAddr string) bool {
	return goos == "linux" && strings.TrimSpace(dbusAddr) == ""
}

func fileKeyringPasswordFunc() keyring.PromptFunc {
	password, passwordSet := os.LookupEnv(keyringPasswordEnv)
	return fileKeyringPasswordFuncFrom(password, passwordSet, term.IsTerminal(int(os.Stdin.Fd())))
}

func fileKeyringPasswordFuncFrom(password string, passwordSet bool, isTTY bool) keyring.PromptFunc {
	// Treat "set to empty string" as intentional; empty passphrase is valid.
	if passwordSet {
		return keyring.FixedStringPrompt(password)
	}

	if isTTY {
		return keyring.TerminalPrompt
	}

	return func(_ string) (string, error) {
		return "", fmt.Errorf("%w; set %s", errNoTTY, keyringPasswordEnv)
	}
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
