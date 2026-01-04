package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeopleGetCmd_HasPersonalFlagWithCorrectDefaults(t *testing.T) {
	// Verify the --personal flag exists on the people get command
	flag := peopleGetCmd.Flags().Lookup("personal")
	require.NotNil(t, flag, "expected --personal flag to exist")
	assert.Equal(t, "false", flag.DefValue)
	assert.Contains(t, flag.Usage, "personal info")
}
