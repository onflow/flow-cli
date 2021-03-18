package accounts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	jsoncdc "github.com/onflow/cadence/encoding/json"

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
	staking := cadenceToJSON(*r.staking)
	delegation := cadenceToJSON(*r.delegation)

	return staking + delegation
}

// String convert result to string
func (r *StakingResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	staking := cadenceToJSON(*r.staking)
	delegation := cadenceToJSON(*r.delegation)

	fmt.Fprint(writer, "Staking Info\n")
	fmt.Fprint(writer, "%v\n\n", staking)

	fmt.Fprint(writer, "Staking Info\n")
	fmt.Fprint(writer, "%v\n", delegation)

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *StakingResult) Oneliner() string {
	return fmt.Sprintf("")
}

func cadenceToJSON(value cadence.Value) string {
	var prettyJSON bytes.Buffer

	b, err := jsoncdc.Encode(value)
	err = json.Indent(&prettyJSON, b, "    ", "    ")
	if err != nil {
		return ""
	}

	return prettyJSON.String()
}
