package accounts

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/spf13/cobra"
)

type flagsStakingInfo struct{}

var StakingCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "staking-info <address>",
		Short: "Get account staking info",
		Args:  cobra.ExactArgs(1),
	},
	Flags: &flagsStakingInfo{},
	Run: func(
		cmd *cobra.Command,
		args []string,
		project *flow.Project,
		services *services.Services,
	) (command.Result, error) {
		staking, delegation, err := services.Accounts.StakingInfo(args[0]) // address
		if err != nil {
			return nil, err
		}

		return &StakingResult{*staking, *delegation}, nil
	},
}

// StakingResult represent result from all account commands
type StakingResult struct {
	staking    cadence.Value
	delegation cadence.Value
}

// JSON convert result to JSON
func (r *StakingResult) JSON() interface{} {
	result := make(map[string]interface{})
	result["staking"] = flow.NewStakingInfoFromValue(r.staking)
	result["delegation"] = flow.NewStakingInfoFromValue(r.delegation)

	return result
}

// String convert result to string
func (r *StakingResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Account Staking Info:\n")

	stakingInfo := flow.NewStakingInfoFromValue(r.staking)

	fmt.Fprintf(writer, "ID: \t %v\n", stakingInfo["id"])
	fmt.Fprintf(writer, "Initial Weight: \t %v\n", stakingInfo["initialWeight"])
	fmt.Fprintf(writer, "Networking Address: \t %v\n", stakingInfo["networkingAddress"])
	fmt.Fprintf(writer, "Networking Key: \t %v\n", stakingInfo["networkingKey"])
	fmt.Fprintf(writer, "Role: \t %v\n", stakingInfo["role"])
	fmt.Fprintf(writer, "Staking Key: \t %v\n", stakingInfo["stakingKey"])
	fmt.Fprintf(writer, "Tokens Committed: \t %v\n", stakingInfo["tokensCommitted"])
	fmt.Fprintf(writer, "Tokens To Unstake: \t %v\n", stakingInfo["tokensRequestedToUnstake"])
	fmt.Fprintf(writer, "Tokens Rewarded: \t %v\n", stakingInfo["tokensRewarded"])
	fmt.Fprintf(writer, "Tokens Staked: \t %v\n", stakingInfo["tokensStaked"])
	fmt.Fprintf(writer, "Tokens Unstaked: \t %v\n", stakingInfo["tokensUnstaked"])
	fmt.Fprintf(writer, "Tokens Unstaking: \t %v\n", stakingInfo["tokensUnstaking"])
	fmt.Fprintf(writer, "Total Tokens Staked: \t %v\n", stakingInfo["totalTokensStaked"])

	delegationStakingInfo := flow.NewStakingInfoFromValue(r.delegation)

	fmt.Fprintf(writer, "\n\nAccount Delegation Info:\n")
	fmt.Fprintf(writer, "ID: \t %v\n", delegationStakingInfo["id"])
	fmt.Fprintf(writer, "Tokens Committed: \t %v\n", delegationStakingInfo["tokensCommitted"])
	fmt.Fprintf(writer, "Tokens To Unstake: \t %v\n", delegationStakingInfo["tokensRequestedToUnstake"])
	fmt.Fprintf(writer, "Tokens Rewarded: \t %v\n", delegationStakingInfo["tokensRewarded"])
	fmt.Fprintf(writer, "Tokens Staked: \t %v\n", delegationStakingInfo["tokensStaked"])
	fmt.Fprintf(writer, "Tokens Unstaked: \t %v\n", delegationStakingInfo["tokensUnstaked"])
	fmt.Fprintf(writer, "Tokens Unstaking: \t %v\n", delegationStakingInfo["tokensUnstaking"])

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *StakingResult) Oneliner() string {
	return ""
}
