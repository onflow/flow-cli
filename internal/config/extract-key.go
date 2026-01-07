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

package config

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/prompt"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsExtractKey struct {
	All bool `default:"false" flag:"all" info:"Extract keys for all accounts with inline keys"`
}

var extractKeyFlags = flagsExtractKey{}

var extractKeyCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "extract-key [account-name]",
		Short: "Extract account private keys to separate key files",
		Long: `Extracts inline private keys from flow.json to separate .pkey files for improved security.

This converts accounts from the inline key format:
  "my-account": { "address": "...", "key": "deadbeef..." }

To the more secure file-based format:
  "my-account": { "address": "...", "key": { "type": "file", "location": "./my-account.pkey" } }

The private key files are automatically added to .gitignore and .cursorignore.`,
		Example: `flow config extract-key my-account
flow config extract-key --all`,
		Args: cobra.MaximumNArgs(1),
	},
	Flags: &extractKeyFlags,
	RunS:  extractKey,
}

func extractKey(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	hexKeyAccounts := findAccountsWithHexKeys(state)

	var accountsToProcess []string
	if len(args) == 1 {
		accountName := args[0]

		_, err := state.Accounts().ByName(accountName)
		if err != nil {
			return nil, fmt.Errorf("account '%s' not found in configuration", accountName)
		}

		if !slices.Contains(hexKeyAccounts, accountName) {
			return nil, fmt.Errorf("account '%s' already uses a file-based key or has an unsupported key type", accountName)
		}
		accountsToProcess = []string{accountName}
	} else if extractKeyFlags.All {
		if len(hexKeyAccounts) == 0 {
			return &result{result: "No accounts with inline keys found. All accounts already use file-based keys."}, nil
		}
		accountsToProcess = hexKeyAccounts
	} else {
		if len(hexKeyAccounts) == 0 {
			return &result{result: "No accounts with inline keys found. All accounts already use file-based keys."}, nil
		}
		options := append(hexKeyAccounts, "all")
		selected, err := prompt.RunSingleSelect(options, "Select an account to extract key (or 'all' for all accounts)")
		if err != nil {
			return nil, err
		}
		if selected == "all" {
			accountsToProcess = hexKeyAccounts
		} else {
			accountsToProcess = []string{selected}
		}
	}

	extractedFiles := make([]string, 0, len(accountsToProcess))
	for _, accountName := range accountsToProcess {
		keyFilePath, err := extractKeyForAccount(state, accountName)
		if err != nil {
			return nil, fmt.Errorf("failed to extract key for '%s': %w", accountName, err)
		}
		extractedFiles = append(extractedFiles, keyFilePath)
		logger.Info(fmt.Sprintf("%s Extracted key for account '%s' to %s", output.SuccessEmoji(), accountName, keyFilePath))
	}

	err := state.SaveEdited(globalFlags.ConfigPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	return &result{
		result: fmt.Sprintf("Successfully extracted keys for %d account(s). Key files added to .gitignore and .cursorignore.", len(accountsToProcess)),
	}, nil
}

// findAccountsWithHexKeys returns account names that have inline hex keys (not file-based keys)
func findAccountsWithHexKeys(state *flowkit.State) []string {
	var hexKeyAccounts []string
	for _, account := range *state.Accounts() {
		// Check if the key is a HexKey (inline key) using type assertion
		if _, isHexKey := account.Key.(*accounts.HexKey); isHexKey {
			hexKeyAccounts = append(hexKeyAccounts, account.Name)
		}
	}
	return hexKeyAccounts
}

func extractKeyForAccount(state *flowkit.State, accountName string) (string, error) {
	account, err := state.Accounts().ByName(accountName)
	if err != nil {
		return "", fmt.Errorf("account '%s' not found", accountName)
	}

	privateKey, err := account.Key.PrivateKey()
	if err != nil {
		return "", fmt.Errorf("cannot extract key: %w", err)
	}
	if privateKey == nil {
		return "", fmt.Errorf("account '%s' does not have a private key", accountName)
	}

	keyFilePath := accounts.PrivateKeyFile(accountName, "")

	if _, err := state.ReaderWriter().ReadFile(keyFilePath); err == nil {
		return "", fmt.Errorf("key file '%s' already exists. Please remove it first or choose a different account", keyFilePath)
	}

	err = state.ReaderWriter().WriteFile(keyFilePath, []byte((*privateKey).String()), os.FileMode(0600))
	if err != nil {
		return "", fmt.Errorf("failed to write key file: %w", err)
	}

	_ = util.AddToGitIgnore(keyFilePath, state.ReaderWriter())
	_ = util.AddToCursorIgnore(keyFilePath, state.ReaderWriter())

	account.Key = accounts.NewFileKey(
		keyFilePath,
		account.Key.Index(),
		account.Key.SigAlgo(),
		account.Key.HashAlgo(),
		state.ReaderWriter(),
	)

	state.Accounts().AddOrUpdate(account)

	return keyFilePath, nil
}
