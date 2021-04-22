package config

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "config",
	Short:            "Utilities to manage configuration",
	TraverseChildren: true,
}

func init() {
	//InitCommand.AddToParent(Cmd)
	AddCommand.AddToParent(Cmd)
}

// configResult result from configuration
type ConfigResult struct {
	result string
}

// JSON convert result to JSON
func (r *ConfigResult) JSON() interface{} {
	return nil
}

func (r *ConfigResult) String() string {
	if r.result != "" {
		return r.result
	}

	return ""
}

// Oneliner show result as one liner grep friendly
func (r *ConfigResult) Oneliner() string {
	return ""
}
