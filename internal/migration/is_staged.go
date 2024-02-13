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

package migration

import (
	"context"
	"fmt"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
)

var isStagedflags = scripts.Flags{}

var IsStagedCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow is-staged <CONTRACT_NAME> <CONTRACT_ADDRESS>",
		Short:   "checks to see if the contract is staged for migration",
		Example: `flow is-staged HelloWorld 0xhello`,
		Args:    cobra.MinimumNArgs(2),
	},
	Flags: &isStagedflags,
	RunS:  isStaged,
}

func isStaged(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	code, err := state.ReaderWriter().ReadFile(GetStagedContractCodeScriptFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read staged contract code script: %w", err)
	}

	contractName, contractAddress := args[0], args[1]

	caddr := cadence.NewAddress(flowsdk.HexToAddress(contractAddress))

	cname, err := cadence.NewString(contractName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract name: %w", err)
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: code,
			Args: []cadence.Value{caddr, cname},
		},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, err
	}

	return scripts.NewScriptResult(value), nil
}
