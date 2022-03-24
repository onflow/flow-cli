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
	"github.com/onflow/flow-go-sdk"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsGet struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
}

var getFlags = flagsGet{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get <address>",
		Short:   "Gets an account by address",
		Example: "flow accounts get f8d6e0586b0a20c7",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &getFlags,
	Run:   get,
}

func get(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	address := flow.HexToAddress(args[0])

	account, err := services.Accounts.Get(address)
	if err != nil {
		return nil, err
	}

	return &AccountResult{
		Account: account,
		include: getFlags.Include,
	}, nil
}
