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
	"strconv"

	"github.com/onflow/go-ethereum/common/math"

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/onflow/flowkit/v2/accounts"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
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

var addAccountCommand = &command.Command{
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
	globalFlags command.GlobalFlags,
	_ output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	raw, flagsProvided, err := flagsToAccountData(addAccountFlags)
	if err != nil {
		return nil, err
	}

	if !flagsProvided {
		raw = prompt.NewAccountPrompt()
	}

	key, err := parseKey(raw.Key, raw.SigAlgo)
	if err != nil {
		return nil, err
	}

	index, err := parseKeyIndex(raw.KeyIndex)
	if err != nil {
		return nil, err
	}

	accountKey := config.AccountKey{
		Type:       config.KeyTypeHex,
		Index:      index,
		SigAlgo:    crypto.StringToSignatureAlgorithm(raw.SigAlgo),
		HashAlgo:   crypto.StringToHashAlgorithm(raw.HashAlgo),
		PrivateKey: key,
	}

	account := &config.Account{
		Name:    raw.Name,
		Address: flow.HexToAddress(raw.Address),
		Key:     accountKey,
	}

	hexKey := accounts.NewHexKeyFromPrivateKey(account.Key.Index, account.Key.HashAlgo, account.Key.PrivateKey)
	state.Accounts().AddOrUpdate(&accounts.Account{
		Name:    account.Name,
		Address: account.Address,
		Key:     hexKey,
	})

	err = state.SaveEdited(globalFlags.ConfigPaths)
	if err != nil {
		return nil, err
	}

	return &result{
		result: fmt.Sprintf("Account %s added to the configuration", raw.Name),
	}, nil

}

func parseKey(key string, sigAlgo string) (crypto.PrivateKey, error) {
	privateKey, err := crypto.DecodePrivateKeyHex(
		crypto.StringToSignatureAlgorithm(sigAlgo),
		key,
	)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func parseKeyIndex(value string) (uint32, error) {
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid index, must be a number")
	}
	if v < 0 {
		return 0, fmt.Errorf("invalid index, must be positive")
	}
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("invalid index, must be less than %d", math.MaxUint32)
	}

	return uint32(v), nil
}

func flagsToAccountData(flags flagsAddAccount) (*prompt.AccountData, bool, error) {
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

	if flow.HexToAddress(flags.Address) == flow.EmptyAddress {
		return nil, true, fmt.Errorf("invalid address")
	}

	return &prompt.AccountData{
		Name:     flags.Name,
		Address:  flags.Address,
		SigAlgo:  flags.SigAlgo,
		HashAlgo: flags.HashAlgo,
		Key:      flags.Key,
		KeyIndex: flags.KeyIndex,
	}, true, nil
}
