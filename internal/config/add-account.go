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

package config

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsAddAccount struct {
	Name     string `flag:"name" info:"Name for the account"`
	Address  string `flag:"address" info:"Account address"`
	KeyIndex string `default:"0" flag:"key-index" info:"Account key index"`
	SigAlgo  string `default:"ECDSA_P256" flag:"sig-algo" info:"Account key signature algorithm"`
	HashAlgo string `default:"SHA3_256" flag:"hash-algo" info:"Account hash used for the digest"`
	Key      string `flag:"key" info:"Account private key"`
}

var addAccountFlags = flagsAddAccount{}

var AddAccountCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "account",
		Short:   "Add account to configuration",
		Example: "flow config add account",
		Args:    cobra.NoArgs,
	},
	Flags: &addAccountFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		p, err := project.Load(globalFlags.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("configuration does not exists")
		}

		accountData, flagsProvided, err := flagsToAccountData(addAccountFlags)
		if err != nil {
			return nil, err
		}

		if !flagsProvided {
			accountData = output.NewAccountPrompt()
		}

		account, err := config.StringToAccount(
			accountData["name"],
			accountData["address"],
			accountData["keyIndex"],
			accountData["sigAlgo"],
			accountData["hashAlgo"],
			accountData["key"],
		)
		if err != nil {
			return nil, err
		}

		acc, err := project.AccountFromConfig(*account)
		if err != nil {
			return nil, err
		}

		p.AddOrUpdateAccount(acc)

		err = p.SaveDefault()
		if err != nil {
			return nil, err
		}

		return &ConfigResult{
			result: fmt.Sprintf("Account %s added to the configuration", accountData["name"]),
		}, nil

	},
}

func init() {
	AddAccountCommand.AddToParent(AddCmd)
}

func flagsToAccountData(flags flagsAddAccount) (map[string]string, bool, error) {
	if flags.Name == "" && flags.Address == "" && flags.Key == "" {
		return nil, false, nil
	}

	if flags.Name == "" {
		return nil, true, fmt.Errorf("name must be provided")
	} else if flags.Address == "" {
		return nil, true, fmt.Errorf("address must be provided")
	} else if flags.Key == "" {
		return nil, true, fmt.Errorf("key must be provided")
	}

	_, err := config.StringToAddress(flags.Address)
	if err != nil {
		return nil, true, err
	}

	return map[string]string{
		"name":     flags.Name,
		"address":  flags.Address,
		"keyIndex": flags.KeyIndex,
		"sigAlgo":  flags.SigAlgo,
		"hashAlgo": flags.HashAlgo,
		"key":      flags.Key,
	}, true, nil
}
