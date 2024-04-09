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

package migrate

import (
	"context"
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/contract-updater/lib/go/templates"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
)

var getStagedCodeflags struct{}

var getStagedCodeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "staged-code <CONTRACT_NAME>",
		Short:   "returns back the staged code for a contract",
		Example: `flow migrate staged-code HelloWorld`,
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
	err := checkNetwork(flow.Network())
	if err != nil {
		return nil, err
	}

	contractName := args[0]
	addr, err := getAddressByContractName(state, contractName, flow.Network())
	if err != nil {
		return nil, fmt.Errorf("error getting address by contract name: %w", err)
	}

	cName, err := cadence.NewString(contractName)
	if err != nil {
		return nil, fmt.Errorf("error creating cadence string: %w", err)
	}

	caddr := cadence.NewAddress(addr)

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: templates.GenerateGetStagedContractCodeScript(MigrationContractStagingAddress(flow.Network().Name)),
			Args: []cadence.Value{caddr, cName},
		},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, fmt.Errorf("error executing script: %w", err)
	}

	return scripts.NewScriptResult(value), nil
}
