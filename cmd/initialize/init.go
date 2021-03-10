package initialize

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsInit struct {
}

type cmdGet struct {
	cmd   *cobra.Command
	flags flagsInit
}

// NewExecuteScriptCmd creates new script command
func NewInitCmd() cmd.Command {
	return &cmdGet{
		cmd: &cobra.Command{
			Use:   "init",
			Short: "Initialize a new account profile",
			Args:  cobra.ExactArgs(1),
		},
	}
}

// Run script command
func (s *cmdGet) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {

}

// GetFlags for script
func (s *cmdGet) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd get command
func (s *cmdGet) GetCmd() *cobra.Command {
	return s.cmd
}

type InitResult struct{}

// JSON convert result to JSON
func (r *InitResult) JSON() interface{} {
	return r
}

// String convert result to string
func (r *InitResult) String() string {
	return ""
}

// Oneliner show result as one liner grep friendly
func (r *InitResult) Oneliner() string {
	return ""
}
