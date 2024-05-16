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

package util

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
)

const EnvPrefix = "FLOW"

func Exit(code int, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}

// AddToGitIgnore adds a new line to the .gitignore if one doesn't exist it creates it.
func AddToGitIgnore(filename string, loader flowkit.ReaderWriter) error {
	currentWd, err := os.Getwd()
	if err != nil {
		return err
	}
	gitIgnorePath := filepath.Join(currentWd, ".gitignore")
	gitIgnoreFiles := ""
	filePermissions := os.FileMode(0644)

	fileStat, err := os.Stat(gitIgnorePath)
	if !os.IsNotExist(err) { // if gitignore exists
		gitIgnoreFilesRaw, err := loader.ReadFile(gitIgnorePath)
		if err != nil {
			return err
		}
		gitIgnoreFiles = string(gitIgnoreFilesRaw)
		filePermissions = fileStat.Mode().Perm()
	}
	return loader.WriteFile(
		gitIgnorePath,
		[]byte(fmt.Sprintf("%s\n%s", gitIgnoreFiles, filename)),
		filePermissions,
	)
}

// GetAddressNetwork returns the chain ID for an address.
func GetAddressNetwork(address flowsdk.Address) (flowsdk.ChainID, error) {
	networks := []flowsdk.ChainID{
		flowsdk.Mainnet,
		flowsdk.Testnet,
		flowsdk.Emulator,
		flowsdk.Previewnet,
	}
	for _, net := range networks {
		if address.IsValid(net) {
			return net, nil
		}
	}

	return "", fmt.Errorf("address not valid for any known chain: %s", address)
}

func CreateTabWriter(b *bytes.Buffer) *tabwriter.Writer {
	return tabwriter.NewWriter(b, 0, 8, 1, '\t', tabwriter.AlignRight)
}

// ValidateECDSAP256Pub attempt to decode the hex string representation of a ECDSA P256 public key
func ValidateECDSAP256Pub(key string) error {
	b, err := hex.DecodeString(strings.TrimPrefix(key, "0x"))
	if err != nil {
		return fmt.Errorf("failed to decode public key hex string: %w", err)
	}

	_, err = crypto.DecodePublicKey(crypto.ECDSA_P256, b)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	return nil
}

func removeFromStringArray(s []string, el string) []string {
	for i, v := range s {
		if v == el {
			s = append(s[:i], s[i+1:]...)
			break
		}
	}

	return s
}

func GetAccountByContractName(state *flowkit.State, contractName string, network config.Network) (*accounts.Account, error) {
	deployments := state.Deployments().ByNetwork(network.Name)
	var accountName string
	for _, d := range deployments {
		for _, c := range d.Contracts {
			if c.Name == contractName {
				accountName = d.Account
				break
			}
		}
	}
	if accountName == "" {
		return nil, fmt.Errorf("contract not found in state")
	}

	accs := state.Accounts()
	if accs == nil {
		return nil, fmt.Errorf("no accounts found in state")
	}

	var account *accounts.Account
	for _, a := range *accs {
		if accountName == a.Name {
			account = &a
			break
		}
	}
	if account == nil {
		return nil, fmt.Errorf("account %s not found in state", accountName)
	}

	return account, nil
}

func GetAddressByContractName(state *flowkit.State, contractName string, network config.Network) (flow.Address, error) {
	account, err := GetAccountByContractName(state, contractName, network)
	if err != nil {
		return flow.Address{}, err
	}

	return flow.HexToAddress(account.Address.Hex()), nil
}

func CheckNetwork(network config.Network) error {
	if network.Name != config.TestnetNetwork.Name && network.Name != config.MainnetNetwork.Name {
		return fmt.Errorf("staging contracts is only supported on testnet & mainnet networks, see https://cadence-lang.org/docs/cadence-migration-guide for more information")
	}
	return nil
}

func NormalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

func Pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}
