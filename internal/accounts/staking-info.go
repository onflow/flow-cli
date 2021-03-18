package accounts

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsStakingInfo struct{}

type cmdStakingInfo struct {
	cmd   *cobra.Command
	flags flagsStakingInfo
}

func NewStakingInfoCmd() command.Command {
	return &cmdStakingInfo{
		cmd: &cobra.Command{
			Use:   "staking-info <address>",
			Short: "Get account staking info",
			Args:  cobra.ExactArgs(1),
		},
	}
}

func (a *cmdStakingInfo) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {
	staking, delegation, err := services.Accounts.StakingInfo(args[0])
	return &StakingResult{staking, delegation}, err
}

func (a *cmdStakingInfo) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdStakingInfo) GetCmd() *cobra.Command {
	return a.cmd
}

// AccountResult represent result from all account commands
type StakingResult struct {
	staking    *cadence.Value
	delegation *cadence.Value
}

// JSON convert result to JSON
func (r *StakingResult) JSON() interface{} {
	result := make(map[string]interface{}, 0)
	result["staking"] = *r.staking
	result["delegation"] = *r.delegation

	return result
}

// String convert result to string
func (r *StakingResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Account Staking Info:\n")
	fmt.Fprintf(writer, "%v\n\n", *r.staking)

	fmt.Fprintf(writer, "Account Delegation Info:\n")
	fmt.Fprintf(writer, "%v\n", *r.delegation)

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *StakingResult) Oneliner() string {
	return fmt.Sprintf("")
}
