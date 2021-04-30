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
	result["staking"] = flowcli.NewStakingInfoFromValue(r.staking)
	result["delegation"] = flowcli.NewStakingInfoFromValue(r.delegation)

	return result
}

// String convert result to string
func (r *StakingResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	fmt.Fprintf(writer, "Account Staking Info:\n")

	stakingInfo := flowcli.NewStakingInfoFromValue(r.staking)

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

	delegationStakingInfo := flowcli.NewStakingInfoFromValue(r.delegation)

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
