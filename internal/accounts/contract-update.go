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
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

var updateContractFlags = deployContractFlags{}

var UpdateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "update-contract <filename> <args>",
		Short:   "Update a contract deployed to an account",
		Example: `flow accounts update-contract ./FungibleToken.cdc helloArg`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &updateContractFlags,
	RunS:  deployContract(false, &updateContractFlags),
}
