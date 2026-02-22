package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOpenClawEnv_MissingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, LoadOpenClawEnv())
}

func TestLoadOpenClawEnv_LoadsUnsetVariablesOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DEELCLI_TEST_EXISTING", "already-set")

	envDir := filepath.Join(home, ".openclaw")
	require.NoError(t, os.MkdirAll(envDir, 0o700))

	envFile := filepath.Join(envDir, ".env")
	content := `
# comment
DEELCLI_TEST_FROM_FILE=from-file
DEELCLI_TEST_EXISTING=from-file
export DEELCLI_TEST_EXPORT=exported
DEELCLI_TEST_QUOTED="hello world"
DEELCLI_TEST_SINGLE='single value'
DEELCLI_TEST_EMPTY=
`
	require.NoError(t, os.WriteFile(envFile, []byte(content), 0o600))

	require.NoError(t, LoadOpenClawEnv())
	assert.Equal(t, "from-file", os.Getenv("DEELCLI_TEST_FROM_FILE"))
	assert.Equal(t, "already-set", os.Getenv("DEELCLI_TEST_EXISTING"))
	assert.Equal(t, "exported", os.Getenv("DEELCLI_TEST_EXPORT"))
	assert.Equal(t, "hello world", os.Getenv("DEELCLI_TEST_QUOTED"))
	assert.Equal(t, "single value", os.Getenv("DEELCLI_TEST_SINGLE"))
	assert.Equal(t, "", os.Getenv("DEELCLI_TEST_EMPTY"))
}

func TestLoadOpenClawEnv_InvalidLineIsIgnored(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	envDir := filepath.Join(home, ".openclaw")
	require.NoError(t, os.MkdirAll(envDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(envDir, ".env"), []byte("INVALID_LINE\nDEELCLI_TEST_VALID=ok\n"), 0o600))

	require.NoError(t, LoadOpenClawEnv())
	assert.Equal(t, "ok", os.Getenv("DEELCLI_TEST_VALID"))
}
