package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/salmonumbrella/deel-cli/internal/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	args := os.Args[1:]
	agentMode := cmd.IsAgentMode(args)

	if err := cmd.ExecuteContext(ctx, args); err != nil {
		exitCode := cmd.ExitCode(err)

		if agentMode {
			// If no structured error was emitted, fall back to a minimal JSON error object.
			if !cmd.AgentErrorEmitted() {
				payload := map[string]any{
					"ok": false,
					"error": map[string]any{
						"message": err.Error(),
					},
				}
				if b, mErr := json.Marshal(payload); mErr == nil {
					_, _ = os.Stdout.Write(append(b, '\n'))
				}
			}
			os.Exit(exitCode)
		}

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitCode)
	}
}
