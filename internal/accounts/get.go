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

type flagsGet struct {
	Code bool `default:"false" flag:"code" info:"Display code deployed to the account"`
}

type cmdGet struct {
	cmd   *cobra.Command
	flags flagsGet
}

func NewGetCmd() command.Command {
	return &cmdGet{
		cmd: &cobra.Command{
			Use:     "get <address>",
			Short:   "Gets an account by address",
			Aliases: []string{"fetch", "g"},
			Long:    `Gets an account by address (address, balance, keys, code)`,
			Args:    cobra.ExactArgs(1),
		},
	}
}

func (a *cmdGet) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {

	account, err := services.Accounts.Get(args[0])
	return &AccountResult{
		Account:  account,
		showCode: a.flags.Code,
	}, err
}

func (a *cmdGet) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdGet) GetCmd() *cobra.Command {
	return a.cmd
}
