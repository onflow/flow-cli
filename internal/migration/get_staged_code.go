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

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"

	"github.com/onflow/contract-updater/lib/go/templates"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
)

var getStagedCodeflags = scripts.Flags{}

var getStagedCodeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow staged-code <CONTRACT_ADDRESS>",
		Short:   "returns back the staged code for a contract",
		Example: `flow staged-code 0xhello`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &getStagedCodeflags,
	RunS:  getStagedCode,
}

func getStagedCode(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	contractAddress := args[0]

	caddr := cadence.NewAddress(flowsdk.HexToAddress(contractAddress))

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: templates.GenerateGetStagedContractCodeScript(MigrationContractStagingAddress(globalFlags.Network)),
			Args: []cadence.Value{caddr},
		},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, err
	}

	return scripts.NewScriptResult(value), nil
}
