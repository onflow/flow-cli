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
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsAddContract struct {
	Account string `default:"emulator-account" flag:"account,a"`
}

type cmdAddContract struct {
	cmd   *cobra.Command
	flags flagsAddContract
}

func NewAddContractCmd() command.Command {
	return &cmdAddContract{
		cmd: &cobra.Command{
			Use:     "add-contract <name> <filename>",
			Short:   "Deploy a new contract to an account",
			Example: `flow accounts add-contract FungibleToken ./FungibleToken.cdc`,
			Args:    cobra.ExactArgs(2),
		},
	}
}

func (a *cmdAddContract) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {

	account, err := services.Accounts.AddContract(a.flags.Account, args[0], args[1], false)
	return &AccountResult{
		Account:  account,
		showCode: false,
	}, err
}

func (a *cmdAddContract) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdAddContract) GetCmd() *cobra.Command {
	return a.cmd
}
