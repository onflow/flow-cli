package config

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "config",
	Short:            "Manage configuration",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(CompletionCmd)
}

// ConfigResult
type ConfigResult struct {
}

// JSON convert result to JSON
func (c *ConfigResult) JSON() interface{} {
	return nil
}

// String convert result to string
func (c *ConfigResult) String() string {
	return ""
}

// Oneliner show result as one liner grep friendly
func (c *ConfigResult) Oneliner() string {
	return ""
}
