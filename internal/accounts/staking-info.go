/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"context"
	"fmt"

	"github.com/onflow/cadence"
	tmpl "github.com/onflow/flow-core-contracts/lib/go/templates"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsStakingInfo struct{}

var stakingFlags = flagsStakingInfo{}

var stakingCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "staking-info <address>",
		Short:   "Get account staking info",
		Example: "flow accounts staking-info f8d6e0586b0a20c7",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &stakingFlags,
	Run:   stakingInfo,
}

func stakingInfo(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	address := flowsdk.HexToAddress(args[0])

	logger.StartProgress(fmt.Sprintf("Fetching info for %s...", address.String()))
	defer logger.StopProgress()

	cadenceAddress := []cadence.Value{cadence.NewAddress(address)}

	chain, err := util.GetAddressNetwork(address)
	if err != nil {
		return nil, fmt.Errorf("failed to determine network from address, check the address and network")
	}

	if chain == flowsdk.Emulator {
		return nil, fmt.Errorf("emulator chain not supported")
	}

	env := envFromNetwork(chain)

	stakingInfoScript := tmpl.GenerateCollectionGetAllNodeInfoScript(env)
	delegationInfoScript := tmpl.GenerateCollectionGetAllDelegatorInfoScript(env)

	stakingValue, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{Code: stakingInfoScript, Args: cadenceAddress},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting staking info: %s", err.Error())
	}

	delegationValue, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{Code: delegationInfoScript, Args: cadenceAddress},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting delegation info: %s", err.Error())
	}

	// get staking infos and delegation infos
	staking, err := newStakingInfoFromValue(stakingValue)
	if err != nil {
		return nil, fmt.Errorf("error parsing staking info: %s", err.Error())
	}
	delegation, err := newStakingInfoFromValue(delegationValue)
	if err != nil {
		return nil, fmt.Errorf("error parsing delegation info: %s", err.Error())
	}

	// get a set of node ids from all staking infos
	nodeStakes := make(map[string]cadence.Value)
	for _, stakingInfo := range staking {
		nodeID, ok := stakingInfo["id"]
		if ok {
			nodeStakes[nodeIDToString(nodeID)] = nil
		}
	}
	totalCommitmentScript := tmpl.GenerateGetTotalCommitmentBalanceScript(env)

	// foreach node id, get the node total stake
	for nodeID := range nodeStakes {
		stake, err := flow.ExecuteScript(
			context.Background(),
			flowkit.Script{
				Code: totalCommitmentScript,
				Args: []cadence.Value{cadence.String(nodeID)},
			},
			flowkit.LatestScriptQuery,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting total stake for node: %s", err.Error())
		}

		nodeStakes[nodeID] = stake
	}

	// foreach staking info, add the node total stake
	for _, stakingInfo := range staking {
		nodeID, ok := stakingInfo["id"]
		if ok {
			stakingInfo["nodeTotalStake"] = nodeStakes[nodeIDToString(nodeID)].(cadence.UFix64)
		}
	}

	logger.StopProgress()

	return &stakingResult{staking, delegation}, nil
}

func envFromNetwork(network flowsdk.ChainID) tmpl.Environment {
	if network == flowsdk.Mainnet {
		return tmpl.Environment{
			IDTableAddress:       "8624b52f9ddcd04a",
			FungibleTokenAddress: "f233dcee88fe0abe",
			FlowTokenAddress:     "1654653399040a61",
			LockedTokensAddress:  "8d0e87b65159ae63",
			StakingProxyAddress:  "62430cf28c26d095",
		}
	}

	if network == flowsdk.Testnet {
		return tmpl.Environment{
			IDTableAddress:       "9eca2b38b18b5dfe",
			FungibleTokenAddress: "9a0766d93b6608b7",
			FlowTokenAddress:     "7e60df042a9c0868",
			LockedTokensAddress:  "95e019a17d0e23d7",
			StakingProxyAddress:  "7aad92e5a0715d21",
		}
	}

	return tmpl.Environment{}
}

func nodeIDToString(value any) string {
	return string(value.(cadence.String))
}

func newStakingInfoFromValue(value cadence.Value) ([]map[string]cadence.Value, error) {
	stakingInfo := make([]map[string]cadence.Value, 0)
	arrayValue, ok := value.(cadence.Array)
	if !ok {
		return stakingInfo, fmt.Errorf("staking info must be a cadence array")
	}

	if len(arrayValue.Values) == 0 {
		return stakingInfo, nil
	}

	for _, v := range arrayValue.Values {
		vs, ok := v.(cadence.Struct)
		if !ok {
			return stakingInfo, fmt.Errorf("staking info must be a cadence array of structs")
		}

		values := cadence.FieldsMappedByName(vs)
		stakingInfo = append(stakingInfo, values)
	}

	return stakingInfo, nil
}

type stakingResult struct {
	staking    []map[string]cadence.Value // stake as FlowIDTableStaking.NodeInfo
	delegation []map[string]cadence.Value // delegation as FlowIDTableStaking.DelegatorInfo
}

func (r *stakingResult) JSON() any {
	result := make(map[string]any)
	result["staking"] = r.staking
	result["delegation"] = r.delegation

	return result
}

func (r *stakingResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	if len(r.staking) != 0 {
		_, _ = fmt.Fprintf(writer, "Account staking info:\n")

		for _, stakingInfo := range r.staking {
			_, _ = fmt.Fprintf(writer, "\tID: \t %v\n", stakingInfo["id"])
			_, _ = fmt.Fprintf(writer, "\tInitial Weight: \t %v\n", stakingInfo["initialWeight"])
			_, _ = fmt.Fprintf(writer, "\tNetworking Address: \t %v\n", stakingInfo["networkingAddress"])
			_, _ = fmt.Fprintf(writer, "\tNetworking Key: \t %v\n", stakingInfo["networkingKey"])
			_, _ = fmt.Fprintf(writer, "\tRole: \t %v\n", stakingInfo["role"])
			_, _ = fmt.Fprintf(writer, "\tStaking Key: \t %v\n", stakingInfo["stakingKey"])
			_, _ = fmt.Fprintf(writer, "\tTokens Committed: \t %v\n", stakingInfo["tokensCommitted"])
			_, _ = fmt.Fprintf(writer, "\tTokens To Unstake: \t %v\n", stakingInfo["tokensRequestedToUnstake"])
			_, _ = fmt.Fprintf(writer, "\tTokens Rewarded: \t %v\n", stakingInfo["tokensRewarded"])
			_, _ = fmt.Fprintf(writer, "\tTokens Staked: \t %v\n", stakingInfo["tokensStaked"])
			_, _ = fmt.Fprintf(writer, "\tTokens Unstaked: \t %v\n", stakingInfo["tokensUnstaked"])
			_, _ = fmt.Fprintf(writer, "\tTokens Unstaking: \t %v\n", stakingInfo["tokensUnstaking"])
			_, _ = fmt.Fprintf(writer, "\tNode Total Stake (including delegators): \t %v\n", stakingInfo["nodeTotalStake"])
			_, _ = fmt.Fprintf(writer, "\n")
		}
	} else {
		_, _ = fmt.Fprintf(writer, "Account has no stakes.\n")
	}

	if len(r.delegation) != 0 {
		_, _ = fmt.Fprintf(writer, "\nAccount delegation info:\n")

		for _, delegationStakingInfo := range r.delegation {
			_, _ = fmt.Fprintf(writer, "\tID: \t %v\n", delegationStakingInfo["id"])
			_, _ = fmt.Fprintf(writer, "\tNode ID: \t %v\n", delegationStakingInfo["nodeID"])
			_, _ = fmt.Fprintf(writer, "\tTokens Committed: \t %v\n", delegationStakingInfo["tokensCommitted"])
			_, _ = fmt.Fprintf(writer, "\tTokens To Unstake: \t %v\n", delegationStakingInfo["tokensRequestedToUnstake"])
			_, _ = fmt.Fprintf(writer, "\tTokens Rewarded: \t %v\n", delegationStakingInfo["tokensRewarded"])
			_, _ = fmt.Fprintf(writer, "\tTokens Staked: \t %v\n", delegationStakingInfo["tokensStaked"])
			_, _ = fmt.Fprintf(writer, "\tTokens Unstaked: \t %v\n", delegationStakingInfo["tokensUnstaked"])
			_, _ = fmt.Fprintf(writer, "\tTokens Unstaking: \t %v\n", delegationStakingInfo["tokensUnstaking"])
			_, _ = fmt.Fprintf(writer, "\n")
		}
	} else {
		_, _ = fmt.Fprintf(writer, "Account has no delegations.\n")
	}

	writer.Flush()
	return b.String()
}

func (r *stakingResult) Oneliner() string {
	return ""
}
