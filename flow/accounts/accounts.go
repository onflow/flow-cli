/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

	add_contract "github.com/onflow/flow-cli/flow/accounts/add-contract"
	"github.com/onflow/flow-cli/flow/accounts/create"
	"github.com/onflow/flow-cli/flow/accounts/get"
	staking_info "github.com/onflow/flow-cli/flow/accounts/staking-info"
	update_contract "github.com/onflow/flow-cli/flow/accounts/update-contract"
)

var Cmd = &cobra.Command{
	Use:              "accounts",
	Short:            "Utilities to manage accounts",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(create.Cmd)
	Cmd.AddCommand(get.Cmd)
	Cmd.AddCommand(staking_info.Cmd)
	Cmd.AddCommand(add_contract.Cmd)
	Cmd.AddCommand(update_contract.Cmd)
}
