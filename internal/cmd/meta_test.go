package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkCommands_ReturnsNonEmpty(t *testing.T) {
	cmds := walkCommands(rootCmd)
	assert.NotEmpty(t, cmds, "walkCommands should return at least the root command")
}

func TestBuildCommandInfo_RootCommand(t *testing.T) {
	info := buildCommandInfo(rootCmd)
	assert.Equal(t, "deel", info.Use)
	assert.NotEmpty(t, info.Short)
	assert.NotEmpty(t, info.Subcommands)
}

func TestFindCommand_ValidPath(t *testing.T) {
	cmd, err := findCommand(rootCmd, []string{"people", "list"})
	require.NoError(t, err)
	assert.Equal(t, "list", cmd.Name())
}

func TestFindCommand_UnknownPath(t *testing.T) {
	_, err := findCommand(rootCmd, []string{"nonexistent", "command"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command path")
}

func TestFindCommand_EmptyPath(t *testing.T) {
	cmd, err := findCommand(rootCmd, []string{})
	require.NoError(t, err)
	assert.Equal(t, rootCmd, cmd)
}

func TestFindCommand_NoProgressLoop(t *testing.T) {
	// cobra's Find returns root when no match is found; the guard should catch this
	_, err := findCommand(rootCmd, []string{"zzz_does_not_exist"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command path")
}
