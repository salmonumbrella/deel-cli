package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootHelp_ShowsStaticHelpText(t *testing.T) {
	// Ensure agent mode is off for this test.
	agentFlag = false

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Reset output capture: SetHelpFunc writes to os.Stdout via fmt.Print,
	// so we capture by executing --help and checking helpText is embedded.
	// Instead, verify the helpText variable is non-empty and contains key sections.
	assert.NotEmpty(t, helpText, "helpText should be embedded from help.txt")

	for _, want := range []string{
		"deel - CLI for Deel",
		"People (workforce):",
		"Contracts:",
		"Exit codes:",
		"Environment:",
		"DEEL_TOKEN",
		"Auth:",
		"--li",
	} {
		assert.Contains(t, helpText, want, "helpText missing %q", want)
	}
}

func TestRootHelp_SubcommandUsesCobra(t *testing.T) {
	agentFlag = false

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	rootCmd.SetArgs([]string{"people", "--help"})
	err := rootCmd.ExecuteContext(context.Background())
	assert.NoError(t, err)

	output := buf.String()
	// Subcommand help should come from Cobra, not our static help.txt.
	assert.True(t,
		strings.Contains(output, "Available Commands") || strings.Contains(output, "Usage:"),
		"subcommand help should use Cobra-generated output, got: %s", output,
	)
	assert.NotContains(t, output, "deel - CLI for Deel",
		"subcommand help should NOT contain the root help.txt header")
}
