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

type flagsRemoveContract struct {
	Signer string `default:"emulator-account" flag:"signer"`
}

type cmdRemoveContract struct {
	cmd   *cobra.Command
	flags flagsRemoveContract
}

func NewRemoveContractCmd() command.Command {
	return &cmdRemoveContract{
		cmd: &cobra.Command{
			Use:     "remove-contract <name>",
			Short:   "Remove a contract deployed to an account",
			Example: `flow accounts remove-contract FungibleToken`,
			Args:    cobra.ExactArgs(1),
		},
	}
}

func (c *cmdRemoveContract) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {
	account, err := services.Accounts.RemoveContract(args[0], c.flags.Signer)
	if err != nil {
		return nil, err
	}

	return &AccountResult{
		Account:  account,
		showCode: false,
	}, nil
}

func (c *cmdRemoveContract) GetFlags() *sconfig.Config {
	return sconfig.New(&c.flags)
}

func (c *cmdRemoveContract) GetCmd() *cobra.Command {
	return c.cmd
}
