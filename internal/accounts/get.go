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
	"fmt"

	"github.com/onflow/flow-go-sdk"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsGet struct {
	Contracts bool     `default:"false" flag:"contracts" info:"⚠️  Deprecated: use include flag instead"`
	Code      bool     `default:"false" flag:"code" info:"⚠️  Deprecated: use contracts flag instead"`
	Include   []string `default:"" flag:"include" info:"Fields to include in the output"`
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
	Run: func(
		cmd *cobra.Command,
		args []string,
		readerWriter flowkit.ReaderWriter,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		if getFlags.Code {
			fmt.Println("⚠️  DEPRECATION WARNING: use contracts flag instead")
		}

		if getFlags.Contracts {
			fmt.Println("⚠️  DEPRECATION WARNING: use include contracts flag instead")
		}

		address := flow.HexToAddress(args[0])

		account, err := services.Accounts.Get(address)
		if err != nil {
			return nil, err
		}

		return &AccountResult{
			Account:  account,
			showCode: getFlags.Contracts || getFlags.Code,
			include:  getFlags.Include,
		}, nil
	},
}
