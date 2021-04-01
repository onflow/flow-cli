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

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsGet struct {
	Contracts bool `default:"false" flag:"contracts" info:"Display contracts deployed to the account"`
	Code      bool `default:"false" flag:"code" info:"⚠️ No longer supported: use contracts flag instead"`
}

var getFlags = flagsGet{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "get <address>",
		Short: "Gets an account by address",
		Args:  cobra.ExactArgs(1),
	},
	Flags: &getFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		if getFlags.Code {
			return nil, fmt.Errorf("⚠️ No longer supported: use contracts flag instead")
		}

		account, err := services.Accounts.Get(args[0]) // address
		if err != nil {
			return nil, err
		}

		return &AccountResult{
			Account:  account,
			showCode: getFlags.Contracts,
		}, nil
	},
}
