package secrets

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deel-cli/internal/config"
)

func newTestStore() *KeyringStore {
	return &KeyringStore{ring: keyring.NewArrayKeyring(nil)}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MyAccount", "myaccount"},
		{"  UPPER  ", "upper"},
		{"already-lower", "already-lower"},
		{"", ""},
		{"  ", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalize(tt.input))
	}
}

func TestCredentialKey(t *testing.T) {
	assert.Equal(t, "account:myaccount", credentialKey("myaccount"))
	assert.Equal(t, "account:", credentialKey(""))
}

func TestParseCredentialKey(t *testing.T) {
	tests := []struct {
		input string
		name  string
		ok    bool
	}{
		{"account:myaccount", "myaccount", true},
		{"account:My Account", "My Account", true},
		{"other:key", "", false},
		{"account:", "", false},
		{"account:  ", "", false},
		{"noprefix", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		name, ok := ParseCredentialKey(tt.input)
		assert.Equal(t, tt.ok, ok, "input: %q", tt.input)
		assert.Equal(t, tt.name, name, "input: %q", tt.input)
	}
}

func TestStore_SetAndGet(t *testing.T) {
	s := newTestStore()
	creds := Credentials{
		Token:     "test-token-123",
		CreatedAt: time.Now().UTC(),
	}

	err := s.Set("MyAccount", creds)
	require.NoError(t, err)

	got, err := s.Get("myaccount")
	require.NoError(t, err)
	assert.Equal(t, "myaccount", got.Name)
	assert.Equal(t, "test-token-123", got.Token)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestStore_SetEmptyName(t *testing.T) {
	s := newTestStore()
	err := s.Set("", Credentials{Token: "token"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing account name")
}

func TestStore_SetEmptyToken(t *testing.T) {
	s := newTestStore()
	err := s.Set("acct", Credentials{Token: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing token")
}

func TestStore_GetEmptyName(t *testing.T) {
	s := newTestStore()
	_, err := s.Get("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing account name")
}

func TestStore_GetNotFound(t *testing.T) {
	s := newTestStore()
	_, err := s.Get("nonexistent")
	assert.Error(t, err)
}

func TestStore_Delete(t *testing.T) {
	s := newTestStore()
	require.NoError(t, s.Set("acct", Credentials{Token: "tok"}))

	err := s.Delete("acct")
	require.NoError(t, err)

	_, err = s.Get("acct")
	assert.Error(t, err)
}

func TestStore_DeleteEmptyName(t *testing.T) {
	s := newTestStore()
	err := s.Delete("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing account name")
}

func TestStore_List(t *testing.T) {
	s := newTestStore()
	require.NoError(t, s.Set("alpha", Credentials{Token: "tok-a"}))
	require.NoError(t, s.Set("beta", Credentials{Token: "tok-b"}))

	list, err := s.List()
	require.NoError(t, err)
	assert.Len(t, list, 2)

	names := make(map[string]bool)
	for _, c := range list {
		names[c.Name] = true
	}
	assert.True(t, names["alpha"])
	assert.True(t, names["beta"])
}

func TestStore_ListEmpty(t *testing.T) {
	s := newTestStore()
	list, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestStore_SetDefaultsCreatedAt(t *testing.T) {
	s := newTestStore()
	before := time.Now().UTC()
	require.NoError(t, s.Set("acct", Credentials{Token: "tok"}))
	after := time.Now().UTC()

	got, err := s.Get("acct")
	require.NoError(t, err)
	assert.False(t, got.CreatedAt.Before(before))
	assert.False(t, got.CreatedAt.After(after))
}

func TestShouldForceFileBackend(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		dbusAddr string
		want     bool
	}{
		{name: "linux headless", goos: "linux", dbusAddr: "", want: true},
		{name: "linux headless whitespace", goos: "linux", dbusAddr: "   ", want: true},
		{name: "linux desktop", goos: "linux", dbusAddr: "unix:path=/run/user/1000/bus", want: false},
		{name: "darwin headless", goos: "darwin", dbusAddr: "", want: false},
	}

	for _, tt := range tests {
		got := shouldForceFileBackend(tt.goos, tt.dbusAddr)
		assert.Equal(t, tt.want, got, tt.name)
	}
}

func TestFileKeyringPasswordFuncFrom_EnvVar(t *testing.T) {
	prompt := fileKeyringPasswordFuncFrom("secret-passphrase", true, false)
	password, err := prompt("ignored")
	require.NoError(t, err)
	assert.Equal(t, "secret-passphrase", password)
}

func TestFileKeyringPasswordFuncFrom_EmptyEnvVar(t *testing.T) {
	prompt := fileKeyringPasswordFuncFrom("", true, false)
	password, err := prompt("ignored")
	require.NoError(t, err)
	assert.Equal(t, "", password)
}

func TestFileKeyringPasswordFuncFrom_NoTTY(t *testing.T) {
	prompt := fileKeyringPasswordFuncFrom("", false, false)
	password, err := prompt("ignored")
	require.Error(t, err)
	assert.Empty(t, password)
	assert.ErrorIs(t, err, errNoTTY)
	assert.Contains(t, err.Error(), keyringPasswordEnv)
}

func TestEnsureKeyringDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg-config"))

	configDir, err := os.UserConfigDir()
	require.NoError(t, err)

	keyringDir, err := ensureKeyringDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(configDir, config.AppName, "keyring"), keyringDir)

	info, err := os.Stat(keyringDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestEnsureKeyringDirPathIsFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg-config"))

	configDir, err := os.UserConfigDir()
	require.NoError(t, err)

	keyringDir := filepath.Join(configDir, config.AppName, "keyring")
	require.NoError(t, os.MkdirAll(filepath.Dir(keyringDir), 0o700))
	require.NoError(t, os.WriteFile(keyringDir, []byte("not-a-directory"), 0o600))

	_, err = ensureKeyringDir()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}
