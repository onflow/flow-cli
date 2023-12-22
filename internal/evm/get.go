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

package evm

import (
	"context"
	_ "embed"
	"encoding/hex"
	"fmt"

	"github.com/onflow/cadence"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/arguments"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
)

//go:embed get.cdc
var getCode []byte

type flagsGet struct{}

var getFlags = flagsGet{}

var getCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get-balance <evm address>",
		Short:   "Get account balance by the EVM address",
		Args:    cobra.ExactArgs(1),
		Example: "flow evm get-balance 522b3294e6d06aa25ad0f1b8891242e335d3b459",
	},
	Flags: &getFlags,
	RunS:  get,
}

// todo only for demo, super hacky now

func get(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	a, err := hex.DecodeString(args[0])
	if err != nil {
		return nil, err
	}

	addressBytes := cadenceByteArrayString(a)

	val, _ := GetEVMAccountBalance(addressBytes, flow)

	fmt.Printf("\n🔥🔥🔥🔥🔥🔥🔥 EVM Get Balance 🔥🔥🔥🔥🔥🔥🔥\n")
	fmt.Println("Balance:  ", val)
	fmt.Printf("\n-------------------------------------------------------------\n\n")
	return nil, nil
}

func GetEVMAccountBalance(
	address string,
	flow flowkit.Services,
) (cadence.Value, error) {

	scriptArgs, err := arguments.ParseWithoutType([]string{address}, getCode, "")
	if err != nil {
		return nil, err
	}

	return flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: getCode,
			Args: scriptArgs,
		},
		flowkit.ScriptQuery{Latest: true},
	)
}
