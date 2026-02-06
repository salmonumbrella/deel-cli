package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/salmonumbrella/deel-cli/internal/outfmt"
)

type flagInfo struct {
	Name         string      `json:"name"`
	Shorthand    string      `json:"shorthand,omitempty"`
	Type         string      `json:"type"`
	DefaultValue string      `json:"default,omitempty"`
	Usage        string      `json:"usage,omitempty"`
	Required     bool        `json:"required,omitempty"`
	Hidden       bool        `json:"hidden,omitempty"`
	Value        interface{} `json:"value,omitempty"`
}

type commandInfo struct {
	Path        string      `json:"path"`
	Use         string      `json:"use"`
	Short       string      `json:"short,omitempty"`
	Long        string      `json:"long,omitempty"`
	Example     string      `json:"example,omitempty"`
	Aliases     []string    `json:"aliases,omitempty"`
	Hidden      bool        `json:"hidden,omitempty"`
	Deprecated  string      `json:"deprecated,omitempty"`
	Flags       []flagInfo  `json:"flags,omitempty"`
	PFlags      []flagInfo  `json:"persistent_flags,omitempty"`
	Subcommands []string    `json:"subcommands,omitempty"`
	ValidArgs   []string    `json:"valid_args,omitempty"`
	Annotations interface{} `json:"annotations,omitempty"`
}

func flagToInfo(f *pflag.Flag) flagInfo {
	fi := flagInfo{
		Name:         f.Name,
		Shorthand:    f.Shorthand,
		Type:         f.Value.Type(),
		DefaultValue: f.DefValue,
		Usage:        f.Usage,
		Hidden:       f.Hidden,
	}
	if f.Annotations != nil {
		if v, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(v) > 0 && v[0] == "true" {
			fi.Required = true
		}
	}
	return fi
}

func buildCommandInfo(c *cobra.Command) commandInfo {
	info := commandInfo{
		Path:       c.CommandPath(),
		Use:        c.Use,
		Short:      c.Short,
		Long:       c.Long,
		Example:    c.Example,
		Aliases:    c.Aliases,
		Hidden:     c.Hidden,
		Deprecated: c.Deprecated,
		ValidArgs:  c.ValidArgs,
	}

	// Persistent flags from this command only (not including parents), plus local flags.
	c.Flags().VisitAll(func(f *pflag.Flag) {
		info.Flags = append(info.Flags, flagToInfo(f))
	})
	c.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		info.PFlags = append(info.PFlags, flagToInfo(f))
	})

	sub := c.Commands()
	for _, sc := range sub {
		if sc.Hidden {
			continue
		}
		info.Subcommands = append(info.Subcommands, sc.Name())
	}
	sort.Strings(info.Subcommands)

	return info
}

func walkCommands(root *cobra.Command) []commandInfo {
	var out []commandInfo
	var visit func(c *cobra.Command)
	visit = func(c *cobra.Command) {
		if !c.Hidden {
			out = append(out, buildCommandInfo(c))
		}
		for _, sc := range c.Commands() {
			visit(sc)
		}
	}
	visit(root)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func findCommand(root *cobra.Command, path []string) (*cobra.Command, error) {
	c := root
	for _, p := range path {
		next, _, err := c.Find([]string{p})
		if err != nil || next == nil {
			return nil, fmt.Errorf("unknown command path: %v", path)
		}
		c = next
	}
	return c, nil
}

var metaCmd = &cobra.Command{
	Use:   "meta",
	Short: "Machine-readable CLI metadata",
	Long:  "Commands in this group output machine-readable metadata for agents/tools.",
}

var metaCommandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "List all commands as JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		if !f.IsJSON() {
			f.PrintText("Use --json or --agent for machine-readable output.")
			return nil
		}
		all := walkCommands(cmd.Root())
		if outfmt.IsAgent(cmd.Context()) {
			return f.PrintJSON(map[string]any{"ok": true, "result": all})
		}
		return f.PrintJSON(all)
	},
}

var metaHelpCmd = &cobra.Command{
	Use:   "help [command path...]",
	Short: "Show command help/schema as JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		if !f.IsJSON() {
			f.PrintText("Use --json or --agent for machine-readable output.")
			return nil
		}
		target, err := findCommand(cmd.Root(), args)
		if err != nil {
			return HandleError(f, err, "resolve command")
		}
		info := buildCommandInfo(target)
		if outfmt.IsAgent(cmd.Context()) {
			return f.PrintJSON(map[string]any{"ok": true, "result": info})
		}
		return f.PrintJSON(info)
	},
}

func init() {
	metaCmd.AddCommand(metaCommandsCmd)
	metaCmd.AddCommand(metaHelpCmd)
	rootCmd.AddCommand(metaCmd)
}
