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
	"github.com/onflow/contract-updater/lib/go/templates"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
)

var isValidatedflags struct{}

var IsValidatedCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "is-validated <CONTRACT_NAME>",
		Short:   "checks to see if the contract has passed the last emulated migration",
		Example: `flow migrate is-validated HelloWorld`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &isValidatedflags,
	RunS:  isValidated,
}

func isValidated(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	contractName := args[0]

	addr, err := getAddressByContractName(state, contractName, flow.Network())
	if err != nil {
		return nil, fmt.Errorf("error getting address by contract name: %w", err)
	}

	caddr := cadence.NewAddress(addr)

	cname, err := cadence.NewString(contractName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract name: %w", err)
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: templates.GenerateIsValidatedScript(MigrationContractStagingAddress(flow.Network().Name)),
			Args: []cadence.Value{caddr, cname},
		},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, fmt.Errorf("error executing script: %w", err)
	}

	return scripts.NewScriptResult(value), nil
}
