package project

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/flow/project/contracts"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsDeploy struct {
	Network string `flag:"network" default:"emulator" info:"network configuration to use"`
	Update  bool   `flag:"update" default:"false" info:"use update flag to update existing contracts"`
}

type cmdDeploy struct {
	cmd   *cobra.Command
	flags flagsDeploy
}

// NewExecuteScriptCmd creates new script command
func NewDeployCmd() cmd.Command {
	return &cmdDeploy{
		cmd: &cobra.Command{
			Use:   "deploy",
			Short: "Deploy Cadence contracts",
		},
	}
}

// Run script command
func (s *cmdDeploy) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {
	c, err := services.Project.Deploy(s.flags.Network, s.flags.Update)
	return &DeployResult{contracts: c, project: project}, err
}

// GetFlags for script
func (s *cmdDeploy) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd get command
func (s *cmdDeploy) GetCmd() *cobra.Command {
	return s.cmd
}

type DeployResult struct {
	contracts []*contracts.Contract
	project   *cli.Project
}

// JSON convert result to JSON
func (r *DeployResult) JSON() interface{} {
	return r
}

// String convert result to string
func (r *DeployResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	for _, contract := range r.contracts {
		fmt.Fprintf(writer, "%s\t0x%s\n", cli.Bold(contract.Name()), contract.Target())
	}

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *DeployResult) Oneliner() string {
	return ""
}
