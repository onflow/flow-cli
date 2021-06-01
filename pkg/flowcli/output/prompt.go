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

package output

import (
	"fmt"
	"net/url"
	"os"

	"github.com/onflow/flow-cli/pkg/flowcli/util"

	"github.com/onflow/flow-cli/pkg/flowcli/config"

	"github.com/gosuri/uilive"
	"github.com/manifoldco/promptui"

	"github.com/onflow/flow-cli/pkg/flowcli/project"
)

func ApproveTransactionPrompt(transaction *project.Transaction) bool {
	writer := uilive.New()
	tx := transaction.FlowTransaction()

	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "ID\t%s\n", tx.ID())
	fmt.Fprintf(writer, "Payer\t%s\n", tx.Payer.Hex())
	fmt.Fprintf(writer, "Authorizers\t%s\n", tx.Authorizers)

	fmt.Fprintf(writer,
		"\nProposal Key:\t\n    Address\t%s\n    Index\t%v\n    Sequence\t%v\n",
		tx.ProposalKey.Address, tx.ProposalKey.KeyIndex, tx.ProposalKey.SequenceNumber,
	)

	if len(tx.PayloadSignatures) == 0 {
		fmt.Fprintf(writer, "\nNo Payload Signatures\n")
	}

	if len(tx.EnvelopeSignatures) == 0 {
		fmt.Fprintf(writer, "\nNo Envelope Signatures\n")
	}

	for i, e := range tx.PayloadSignatures {
		fmt.Fprintf(writer, "\nPayload Signature %v:\n", i)
		fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
		fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
	}

	for i, e := range tx.EnvelopeSignatures {
		fmt.Fprintf(writer, "\nEnvelope Signature %v:\n", i)
		fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
		fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
	}

	if tx.Script != nil {
		if len(tx.Arguments) == 0 {
			fmt.Fprintf(writer, "\n\nArguments\tNo arguments\n")
		} else {
			fmt.Fprintf(writer, "\n\nArguments (%d):\n", len(tx.Arguments))
			for i, argument := range tx.Arguments {
				fmt.Fprintf(writer, "    - Argument %d: %s\n", i, argument)
			}
		}

		fmt.Fprintf(writer, "\nCode\n\n%s\n", tx.Script)
	}

	fmt.Fprintf(writer, "\n\n")
	writer.Flush()

	prompt := promptui.Select{
		Label: "⚠️  Do you want to sign this transaction?",
		Items: []string{"No", "Yes"},
	}

	_, result, _ := prompt.Run()

	fmt.Fprintf(writer, "\r\r")
	writer.Flush()

	return result == "Yes"
}

func AutocompletionPrompt() (string, string) {
	prompt := promptui.Select{
		Label: "❓ Select your shell (you can run 'echo $SHELL' to find out)",
		Items: []string{"bash", "zsh", "powershell"},
	}

	_, shell, _ := prompt.Run()
	os := ""

	switch shell {
	case "bash":
		prompt := promptui.Select{
			Label: "❓ Select operation system",
			Items: []string{"MacOS", "Linux"},
		}
		_, os, _ = prompt.Run()
	case "powershell":
		fmt.Printf(`PowerShell Installation Guide:
PS> flow config setup-completions powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> flow config setup-completions powershell > flow.ps1
# and source this file from your PowerShell profile.
`)
	}

	return shell, os
}

func namePrompt() string {
	namePrompt := promptui.Prompt{
		Label: "Name",
		Validate: func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("invalid name")
			}
			return nil
		},
	}

	name, err := namePrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return name
}

func addressPrompt() string {
	addressPrompt := promptui.Prompt{
		Label: "Address",
		Validate: func(s string) error {
			_, err := config.StringToAddress(s)
			return err
		},
	}

	address, err := addressPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return address
}

func contractPrompt(contractNames []string) string {
	contractPrompt := promptui.Select{
		Label: "Choose contract you wish to deploy",
		Items: contractNames,
	}
	_, contractName, err := contractPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return contractName
}

func addAnotherContractToDeploymentPrompt() bool {
	addContractPrompt := promptui.Select{
		Label: "Do you wish to add another contract for deployment?",
		Items: []string{"No", "Yes"},
	}
	_, addMore, err := addContractPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return addMore == "Yes"
}

func NewAccountPrompt() map[string]string {
	accountData := make(map[string]string)
	var err error

	accountData["name"] = namePrompt()
	accountData["address"] = addressPrompt()

	sigAlgoPrompt := promptui.Select{
		Label: "Signature algorithm",
		Items: []string{"ECDSA_P256", "ECDSA_secp256k1"},
	}
	_, accountData["sigAlgo"], err = sigAlgoPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	hashAlgoPrompt := promptui.Select{
		Label: "Hashing algorithm",
		Items: []string{"SHA3_256", "SHA2_256"},
	}
	_, accountData["hashAlgo"], err = hashAlgoPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	keyPrompt := promptui.Prompt{
		Label: "Private key",
		Validate: func(s string) error {
			_, err := config.StringToHexKey(s, accountData["sigAlgo"])
			return err
		},
	}
	accountData["key"], err = keyPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	keyIndexPrompt := promptui.Prompt{
		Label:   "Key index (Default: 0)",
		Default: "0",
		Validate: func(s string) error {
			_, err := config.StringToKeyIndex(s)
			return err
		},
	}

	accountData["keyIndex"], err = keyIndexPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return accountData
}

func NewContractPrompt() map[string]string {
	contractData := make(map[string]string)
	var err error

	contractData["name"] = namePrompt()

	sourcePrompt := promptui.Prompt{
		Label: "Contract file location",
		Validate: func(s string) error {
			if !config.Exists(s) {
				return fmt.Errorf("contract file doesn't exist: %s", s)
			}

			return nil
		},
	}
	contractData["source"], err = sourcePrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	emulatorAliasPrompt := promptui.Prompt{
		Label: "Enter emulator alias, if exists",
		Validate: func(s string) error {
			if s != "" {
				_, err := config.StringToAddress(s)
				return err
			}

			return nil
		},
	}
	contractData["emulator"], _ = emulatorAliasPrompt.Run()

	testnetAliasPrompt := promptui.Prompt{
		Label: "Enter testnet alias, if exists",
		Validate: func(s string) error {
			if s != "" {
				_, err := config.StringToAddress(s)
				return err
			}

			return nil
		},
	}
	contractData["testnet"], _ = testnetAliasPrompt.Run()

	return contractData
}

func NewNetworkPrompt() map[string]string {
	networkData := make(map[string]string)
	var err error

	networkData["name"] = namePrompt()

	hostPrompt := promptui.Prompt{
		Label: "Enter host location",
		Validate: func(s string) error {
			_, err := url.ParseRequestURI(s)
			return err
		},
	}
	networkData["host"], err = hostPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return networkData
}

func NewDeploymentPrompt(
	networks config.Networks,
	accounts config.Accounts,
	contracts config.Contracts,
) map[string]interface{} {
	deploymentData := make(map[string]interface{})
	var err error

	networkNames := make([]string, 0)
	for _, network := range networks {
		networkNames = append(networkNames, network.Name)
	}

	networkPrompt := promptui.Select{
		Label: "Choose network for deployment",
		Items: networkNames,
	}
	_, deploymentData["network"], err = networkPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	accountNames := make([]string, 0)
	for _, account := range accounts {
		accountNames = append(accountNames, account.Name)
	}

	accountPrompt := promptui.Select{
		Label: "Choose an account to deploy to",
		Items: accountNames,
	}
	_, deploymentData["account"], err = accountPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	contractNames := make([]string, 0)
	for _, contract := range contracts {
		contractNames = append(contractNames, contract.Name)
	}

	deploymentData["contracts"] = make([]string, 0)

	contractName := contractPrompt(contractNames)
	deploymentData["contracts"] = append(deploymentData["contracts"].([]string), contractName)
	contractNames = util.RemoveFromStringArray(contractNames, contractName)

	for addAnotherContractToDeploymentPrompt() {
		contractName := contractPrompt(contractNames)
		deploymentData["contracts"] = append(deploymentData["contracts"].([]string), contractName)
		contractNames = util.RemoveFromStringArray(contractNames, contractName)

		if len(contractNames) == 0 {
			break
		}
	}

	return deploymentData
}

func RemoveAccountPrompt(accounts config.Accounts) string {
	accountNames := make([]string, 0)

	for _, account := range accounts {
		accountNames = append(accountNames, account.Name)
	}

	namePrompt := promptui.Select{
		Label: "Select an account name you wish to remove",
		Items: accountNames,
	}

	_, name, err := namePrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return name
}

func RemoveDeploymentPrompt(deployments config.Deployments) (account string, network string) {
	deploymentNames := make([]string, 0)

	for _, deployment := range deployments {
		contractNames := make([]string, 0)
		for _, c := range deployment.Contracts {
			contractNames = append(contractNames, c.Name)
		}

		deploymentNames = append(
			deploymentNames,
			fmt.Sprintf(
				"Account: %s, Network: %s, Contracts: %s",
				deployment.Account,
				deployment.Network,
				contractNames,
			),
		)
	}

	deployPrompt := promptui.Select{
		Label: "Select deployment you wish to remove",
		Items: deploymentNames,
	}

	index, _, err := deployPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return deployments[index].Account, deployments[index].Network
}

func RemoveContractPrompt(contracts config.Contracts) string {
	contractNames := make([]string, 0)

	for _, contract := range contracts {
		contractNames = append(contractNames, contract.Name)
	}

	contractPrompt := promptui.Select{
		Label: "Select contract you wish to remove",
		Items: contractNames,
	}

	_, name, err := contractPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return name
}

func RemoveNetworkPrompt(networks config.Networks) string {
	networkNames := make([]string, 0)

	for _, network := range networks {
		networkNames = append(networkNames, network.Name)
	}

	networkPrompt := promptui.Select{
		Label: "Select network you wish to remove",
		Items: networkNames,
	}

	_, name, err := networkPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return name
}

func SnapshotPrompt(snapshots []string) string {
	networkPrompt := promptui.Select{
		Label: "Select snapshot",
		Items: snapshots,
	}

	_, name, err := networkPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return name
}
