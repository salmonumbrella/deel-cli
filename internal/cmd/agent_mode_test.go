package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salmonumbrella/deel-cli/internal/config"
)

func TestIsAgentMode_ArgsOverrideEnv(t *testing.T) {
	t.Setenv(config.EnvAgent, "1")
	assert.True(t, IsAgentMode([]string{}))
	assert.False(t, IsAgentMode([]string{"--agent=false"}))
	assert.True(t, IsAgentMode([]string{"--agent"}))
}

func TestIsAgentMode_EnvTruthy(t *testing.T) {
	_ = os.Unsetenv(config.EnvAgent)
	assert.False(t, IsAgentMode([]string{}))

	t.Setenv(config.EnvAgent, "true")
	assert.True(t, IsAgentMode([]string{}))
}
