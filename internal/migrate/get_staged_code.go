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
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/contract-updater/lib/go/templates"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flow-cli/internal/util"
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
	err := util.CheckNetwork(flow.Network())
	if err != nil {
		return nil, err
	}

	contractName := args[0]
	addr, err := util.GetAddressByContractName(state, contractName, flow.Network())
	if err != nil {
		return nil, fmt.Errorf("error getting address by contract name: %w", err)
	}

	location := common.NewAddressLocation(nil, common.Address(addr), contractName)
	code, err := getStagedContractCode(context.Background(), flow, location)
	if err != nil {
		return nil, err
	}

	// If the contract is not staged, return nil
	if code == nil {
		return scripts.NewScriptResult(cadence.NewOptional(nil)), nil
	}

	return scripts.NewScriptResult(cadence.NewOptional(cadence.String(code))), nil
}

func getStagedContractCode(
	ctx context.Context,
	flow flowkit.Services,
	location common.AddressLocation,
) ([]byte, error) {
	cAddr := cadence.BytesToAddress(location.Address.Bytes())
	cName, err := cadence.NewString(location.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract name: %w", err)
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: templates.GenerateGetStagedContractCodeScript(MigrationContractStagingAddress(flow.Network().Name)),
			Args: []cadence.Value{cAddr, cName},
		},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, err
	}

	optValue, ok := value.(cadence.Optional)
	if !ok {
		return nil, fmt.Errorf("invalid script return value type: %T", value)
	}

	if optValue.Value == nil {
		return nil, nil
	}

	strValue, ok := optValue.Value.(cadence.String)
	if !ok {
		return nil, fmt.Errorf("invalid script return value type: %T", value)
	}

	return []byte(strValue), nil
}
