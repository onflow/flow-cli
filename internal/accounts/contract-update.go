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

type flagsUpdateContract struct {
	Account string `default:"emulator-account" flag:"account,a"`
	Host    string `flag:"host" info:"Flow Access API host address"`
}

type cmdUpdateContract struct {
	cmd   *cobra.Command
	flags flagsUpdateContract
}

func NewUpdateContractCmd() command.Command {
	return &cmdUpdateContract{
		cmd: &cobra.Command{
			Use:     "update-contract <name> <filename>",
			Short:   "Update a contract deployed to an account",
			Example: `flow accounts update-contract FungibleToken ./FungibleToken.cdc`,
			Args:    cobra.ExactArgs(2),
		},
	}
}

func (c *cmdUpdateContract) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {

	account, err := services.Accounts.AddContract(c.flags.Account, args[0], args[1], true)
	return &AccountResult{
		Account:  account,
		showCode: true,
	}, err

}

func (c *cmdUpdateContract) GetFlags() *sconfig.Config {
	return sconfig.New(&c.flags)
}

func (c *cmdUpdateContract) GetCmd() *cobra.Command {
	return c.cmd
}
