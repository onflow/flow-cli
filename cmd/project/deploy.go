package project

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/lib/contracts"
	"github.com/onflow/flow-cli/flow/services"
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
	project *lib.Project,
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
	project   *lib.Project
}

// JSON convert result to JSON
func (r *DeployResult) JSON() interface{} {
	result := make(map[string]string, 0)

	for _, contract := range r.contracts {
		result[contract.Name()] = contract.Target().String()
	}

	return result
}

// String convert result to string
func (r *DeployResult) String() string {
	return ""
}

// Oneliner show result as one liner grep friendly
func (r *DeployResult) Oneliner() string {
	return ""
}
