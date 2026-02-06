package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/salmonumbrella/deel-cli/internal/config"
)

func agentEnabledFromEnv() bool {
	v := strings.TrimSpace(os.Getenv(config.EnvAgent))
	if v == "" {
		return false
	}
	// Accept common truthy values: "1", "true", "TRUE", etc.
	b, err := strconv.ParseBool(v)
	if err != nil {
		// If it's set but unparsable, treat as enabled.
		return true
	}
	return b
}

func agentEnabledFromArgs(args []string) (bool, bool) {
	for _, a := range args {
		if a == "--agent" {
			return true, true
		}
		if strings.HasPrefix(a, "--agent=") {
			raw := strings.TrimPrefix(a, "--agent=")
			b, err := strconv.ParseBool(raw)
			if err != nil {
				return true, true
			}
			return b, true
		}
	}
	return false, false
}

// IsAgentMode returns true if agent mode is enabled via args or environment.
// This is used by main() before Cobra executes, so it must be args-based.
func IsAgentMode(args []string) bool {
	if b, ok := agentEnabledFromArgs(args); ok {
		return b
	}
	return agentEnabledFromEnv()
}
