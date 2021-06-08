/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package accounts

import (
	"bytes"
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/cadence"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

type flagsStakingInfo struct{}

var stakingFlags = flagsStakingInfo{}

var StakingCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "staking-info <address>",
		Short:   "Get account staking info",
		Example: "flow accounts staking-info f8d6e0586b0a20c7",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &stakingFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
		proj *project.Project,
	) (command.Result, error) {
		address := flow.HexToAddress(args[0])

		staking, delegation, err := services.Accounts.StakingInfo(address) // address
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
	result["staking"] = flowcli.NewStakingInfoFromValue(r.staking)
	result["delegation"] = flowcli.NewStakingInfoFromValue(r.delegation)

	return result
}

// String convert result to string
func (r *StakingResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Account Staking Info:\n")

	stakingInfo := flowcli.NewStakingInfoFromValue(r.staking)

	_, _ = fmt.Fprintf(writer, "ID: \t %v\n", stakingInfo["id"])
	_, _ = fmt.Fprintf(writer, "Initial Weight: \t %v\n", stakingInfo["initialWeight"])
	_, _ = fmt.Fprintf(writer, "Networking Address: \t %v\n", stakingInfo["networkingAddress"])
	_, _ = fmt.Fprintf(writer, "Networking Key: \t %v\n", stakingInfo["networkingKey"])
	_, _ = fmt.Fprintf(writer, "Role: \t %v\n", stakingInfo["role"])
	_, _ = fmt.Fprintf(writer, "Staking Key: \t %v\n", stakingInfo["stakingKey"])
	_, _ = fmt.Fprintf(writer, "Tokens Committed: \t %v\n", stakingInfo["tokensCommitted"])
	_, _ = fmt.Fprintf(writer, "Tokens To Unstake: \t %v\n", stakingInfo["tokensRequestedToUnstake"])
	_, _ = fmt.Fprintf(writer, "Tokens Rewarded: \t %v\n", stakingInfo["tokensRewarded"])
	_, _ = fmt.Fprintf(writer, "Tokens Staked: \t %v\n", stakingInfo["tokensStaked"])
	_, _ = fmt.Fprintf(writer, "Tokens Unstaked: \t %v\n", stakingInfo["tokensUnstaked"])
	_, _ = fmt.Fprintf(writer, "Tokens Unstaking: \t %v\n", stakingInfo["tokensUnstaking"])
	_, _ = fmt.Fprintf(writer, "Total Tokens Staked: \t %v\n", stakingInfo["totalTokensStaked"])

	delegationStakingInfo := flowcli.NewStakingInfoFromValue(r.delegation)

	_, _ = fmt.Fprintf(writer, "\n\nAccount Delegation Info:\n")
	_, _ = fmt.Fprintf(writer, "ID: \t %v\n", delegationStakingInfo["id"])
	_, _ = fmt.Fprintf(writer, "Tokens Committed: \t %v\n", delegationStakingInfo["tokensCommitted"])
	_, _ = fmt.Fprintf(writer, "Tokens To Unstake: \t %v\n", delegationStakingInfo["tokensRequestedToUnstake"])
	_, _ = fmt.Fprintf(writer, "Tokens Rewarded: \t %v\n", delegationStakingInfo["tokensRewarded"])
	_, _ = fmt.Fprintf(writer, "Tokens Staked: \t %v\n", delegationStakingInfo["tokensStaked"])
	_, _ = fmt.Fprintf(writer, "Tokens Unstaked: \t %v\n", delegationStakingInfo["tokensUnstaked"])
	_, _ = fmt.Fprintf(writer, "Tokens Unstaking: \t %v\n", delegationStakingInfo["tokensUnstaking"])

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *StakingResult) Oneliner() string {
	return ""
}
