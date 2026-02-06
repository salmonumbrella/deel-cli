package cmd

import "sync/atomic"

var agentErrorEmitted atomic.Bool

func markAgentErrorEmitted() {
	agentErrorEmitted.Store(true)
}

// AgentErrorEmitted returns true if this process has already emitted a structured agent error on stdout.
func AgentErrorEmitted() bool {
	return agentErrorEmitted.Load()
}

// resetAgentErrorEmitted clears the emitted flag; used by tests to ensure isolation between runs.
func resetAgentErrorEmitted() {
	agentErrorEmitted.Store(false)
}
