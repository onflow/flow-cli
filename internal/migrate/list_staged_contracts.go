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

var listStagedContractsflags struct{}

var listStagedContractsCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "list-staged <ACCOUNT_NAME>",
		Short:   "returns back the a list of staged contracts given an account name",
		Example: `flow migrate list-staged test-account`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &listStagedContractsflags,
	RunS:  listStagedContracts,
}

func listStagedContracts(
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

	accountName := args[0]
	account, err := state.Accounts().ByName(accountName)
	if err != nil {
		return nil, fmt.Errorf("error getting account by name: %w", err)
	}

	caddr := cadence.NewAddress(account.Address)

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: templates.GenerateGetStagedContractNamesForAddressScript(MigrationContractStagingAddress(flow.Network().Name)),
			Args: []cadence.Value{caddr},
		},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, fmt.Errorf("error executing script: %w", err)
	}

	return scripts.NewScriptResult(value), nil
}
