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
	"github.com/spf13/cobra"
)

type flagsCreate struct {
	Signer    string   `default:"emulator-account" flag:"signer"`
	Keys      []string `flag:"key" info:"Public keys to attach to account"`
	SigAlgo   string   `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm used to generate the keys"`
	HashAlgo  string   `default:"SHA3_256" flag:"hash-algo" info:"Hash used for the digest"`
	Name      string   `default:"default" flag:"name" info:"Name used for saving account"`
	Contracts []string `flag:"contract" info:"Contract to be deployed during account creation. <name:filename>"`
}

var createFlags = &flagsCreate{}

var CreateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "create",
		Short:   "Create a new account",
		Example: `flow accounts create --key d651f1931a2...8745`,
	},
	Flags: &createFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		project *flow.Project,
		services *services.Services,
	) (command.Result, error) {
		account, err := services.Accounts.Create(
			createFlags.Signer,
			createFlags.Keys,
			createFlags.SigAlgo,
			createFlags.HashAlgo,
			createFlags.Contracts,
		)

		if err != nil {
			return nil, err
		}

		return &AccountResult{
			Account: account,
		}, nil
	},
}
