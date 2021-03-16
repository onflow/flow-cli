package scripts

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsScripts struct {
	ArgsJSON string   `default:"" flag:"argsJSON" info:"arguments in JSON-Cadence format"`
	Args     []string `default:"" flag:"arg" info:"argument in Type:Value format"`
}

type cmdExecuteScript struct {
	cmd   *cobra.Command
	flags flagsScripts
}

// NewExecuteScriptCmd creates new script command
func NewExecuteScriptCmd() cmd.Command {
	return &cmdExecuteScript{
		cmd: &cobra.Command{
			Use:     "execute <filename>",
			Short:   "Execute a script",
			Example: `flow scripts execute script.cdc --arg String:"Hello" --arg String:"World"`,
			Args:    cobra.ExactArgs(1),
		},
	}
}

// Run script command
func (s *cmdExecuteScript) Run(
	cmd *cobra.Command,
	args []string,
	project *lib.Project,
	services *services.Services,
) (cmd.Result, error) {
	value, err := services.Scripts.Execute(args[0], s.flags.Args, s.flags.ArgsJSON) // TODO: add support for json args
	return &ScriptResult{value}, err
}

// GetFlags for script
func (s *cmdExecuteScript) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd get command
func (s *cmdExecuteScript) GetCmd() *cobra.Command {
	return s.cmd
}
