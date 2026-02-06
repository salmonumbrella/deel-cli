package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deel-cli/internal/outfmt"
)

func TestRequireForce_AgentModeEmitsJSONAndErrors(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	f := outfmt.New(&out, &errOut, outfmt.FormatJSON, "never")

	c := &cobra.Command{}
	c.SetContext(outfmt.WithAgent(context.Background(), true))

	ok, err := requireForce(c, f, false, "delete", "task", "t1", "deel tasks delete t1 --force")
	require.False(t, ok)
	require.Error(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &payload))
	assert.Equal(t, false, payload["ok"])
}
