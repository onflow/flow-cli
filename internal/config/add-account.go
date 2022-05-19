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

package config

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsAddAccount struct {
	Name     string `flag:"name" info:"Name for the account"`
	Address  string `flag:"address" info:"Account address"`
	KeyIndex string `default:"0" flag:"key-index" info:"Account key index"`
	SigAlgo  string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm of this account key"`
	HashAlgo string `default:"SHA3_256" flag:"hash-algo" info:"Hash algorithm to pair with this account key"`
	Key      string `flag:"private-key" info:"Account private key"`
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
	RunS:  addAccount,
}

func addAccount(
	_ []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
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

	acc := flowkit.Account{}
	acc.SetName(account.Name)
	acc.SetAddress(account.Address)
	acc.SetKey(flowkit.NewHexAccountKeyFromPrivateKey(account.Key.Index, account.Key.HashAlgo, account.Key.PrivateKey))

	state.Accounts().AddOrUpdate(&acc)

	err = state.SaveEdited(globalFlags.ConfigPaths)
	if err != nil {
		return nil, err
	}

	return &Result{
		result: fmt.Sprintf("Account %s added to the configuration", accountData["name"]),
	}, nil

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
