/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go/fvm/systemcontracts"
	flowGo "github.com/onflow/flow-go/model/flow"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
)

const EnvPrefix = "FLOW"

func Exit(code int, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}

// entryExists checks if an entry already exists in the content
func entryExists(content, entry string) bool {
	lines := strings.SplitSeq(strings.TrimSpace(content), "\n")
	for line := range lines {
		if strings.TrimSpace(line) == strings.TrimSpace(entry) {
			return true
		}
	}
	return false
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

	if entryExists(gitIgnoreFiles, filename) {
		return nil // Entry already exists, no need to add
	}

	return loader.WriteFile(
		gitIgnorePath,
		fmt.Appendf(nil, "%s\n%s", gitIgnoreFiles, filename),
		filePermissions,
	)
}

// AddToCursorIgnore adds a new line to the .cursorignore if one doesn't exist it creates it.
func AddToCursorIgnore(filename string, loader flowkit.ReaderWriter) error {
	currentWd, err := os.Getwd()
	if err != nil {
		return err
	}
	cursorIgnorePath := filepath.Join(currentWd, ".cursorignore")
	cursorIgnoreFiles := ""
	filePermissions := os.FileMode(0644)

	fileStat, err := os.Stat(cursorIgnorePath)
	if !os.IsNotExist(err) {
		cursorIgnoreFilesRaw, err := loader.ReadFile(cursorIgnorePath)
		if err != nil {
			return err
		}
		cursorIgnoreFiles = string(cursorIgnoreFilesRaw)
		filePermissions = fileStat.Mode().Perm()
	}

	if entryExists(cursorIgnoreFiles, filename) {
		return nil // Entry already exists, no need to add
	}

	return loader.WriteFile(
		cursorIgnorePath,
		fmt.Appendf(nil, "%s\n%s", cursorIgnoreFiles, filename),
		filePermissions,
	)
}

// addEntriesToIgnoreFile is a helper function that adds entries to an ignore file without duplicates
func addEntriesToIgnoreFile(filePath string, entries []string, loader flowkit.ReaderWriter) error {
	existingContent := ""
	filePermissions := os.FileMode(0644)

	// Try to read existing content using the loader
	existingContentRaw, err := loader.ReadFile(filePath)
	if err == nil {
		existingContent = string(existingContentRaw)
		// Try to get file permissions, but don't fail if we can't
		if stat, err := os.Stat(filePath); err == nil {
			filePermissions = stat.Mode().Perm()
		}
	}

	// Split existing content into lines
	existingLines := strings.Split(strings.TrimSpace(existingContent), "\n")
	existingSet := make(map[string]bool)
	for _, line := range existingLines {
		if strings.TrimSpace(line) != "" {
			existingSet[strings.TrimSpace(line)] = true
		}
	}

	// Add new entries that don't already exist
	var newEntries []string
	for _, entry := range entries {
		if !existingSet[strings.TrimSpace(entry)] {
			newEntries = append(newEntries, entry)
		}
	}

	if len(newEntries) == 0 {
		return nil // All entries already exist
	}

	// Combine existing content with new entries
	content := existingContent
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += strings.Join(newEntries, "\n")

	return loader.WriteFile(filePath, []byte(content), filePermissions)
}

// AddFlowEntriesToGitIgnore adds the standard Flow entries to .gitignore without duplicates
func AddFlowEntriesToGitIgnore(targetDir string, loader flowkit.ReaderWriter) error {
	flowEntries := []string{
		"# flow",
		"emulator-account.pkey",
		"imports",
		".env",
	}

	gitIgnorePath := filepath.Join(targetDir, ".gitignore")
	return addEntriesToIgnoreFile(gitIgnorePath, flowEntries, loader)
}

// AddFlowEntriesToCursorIgnore adds the standard Flow entries to .cursorignore without duplicates
func AddFlowEntriesToCursorIgnore(targetDir string, loader flowkit.ReaderWriter) error {
	flowEntries := []string{
		"# flow",
		"emulator-account.pkey",
		".env",
		"",
		"# Pay attention to imports directory",
		"!imports",
	}

	cursorIgnorePath := filepath.Join(targetDir, ".cursorignore")
	return addEntriesToIgnoreFile(cursorIgnorePath, flowEntries, loader)
}

// GetAddressNetwork returns the chain ID for an address.
func GetAddressNetwork(address flow.Address) (flow.ChainID, error) {
	networks := []flow.ChainID{
		flow.Mainnet,
		flow.Testnet,
		flow.Emulator,
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
			s = slices.Delete(s, i, i+1)
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

func IsCoreContract(contractName string) bool {
	sc := systemcontracts.SystemContractsForChain(flowGo.Emulator)

	for _, coreContract := range sc.All() {
		if coreContract.Name == contractName {
			return true
		}
	}
	return false
}

// ResolveAddressOrAccountName resolves a string that could be either an address or account name
func ResolveAddressOrAccountName(input string, state *flowkit.State) (flow.Address, error) {
	address := flow.HexToAddress(input)

	if address.IsValid(flow.Mainnet) || address.IsValid(flow.Testnet) || address.IsValid(flow.Emulator) {
		return address, nil
	}

	account, err := state.Accounts().ByName(input)
	if err != nil {
		return flow.EmptyAddress, fmt.Errorf("could not find account with name %s", input)
	}

	return account.Address, nil
}
