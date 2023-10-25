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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gosuri/uilive"
	"github.com/manifoldco/promptui"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/output"
)

func ApproveTransactionForSigningPrompt(transaction *flow.Transaction) bool {
	return ApproveTransactionPrompt(transaction, "⚠️  Do you want to SIGN this transaction?")
}

func ApproveTransactionForBuildingPrompt(transaction *flow.Transaction) bool {
	return ApproveTransactionPrompt(transaction, "⚠️  Do you want to BUILD this transaction?")
}

func ApproveTransactionForSendingPrompt(transaction *flow.Transaction) bool {
	return ApproveTransactionPrompt(transaction, "⚠️  Do you want to SEND this transaction?")
}

func ApproveTransactionPrompt(tx *flow.Transaction, promptMsg string) bool {
	writer := uilive.New()

	_, _ = fmt.Fprintf(writer, "\n")
	_, _ = fmt.Fprintf(writer, "ID\t%s\n", tx.ID())
	_, _ = fmt.Fprintf(writer, "Payer\t%s\n", tx.Payer.Hex())
	_, _ = fmt.Fprintf(writer, "Authorizers\t%s\n", tx.Authorizers)

	_, _ = fmt.Fprintf(writer,
		"\nProposal Key:\t\n    Address\t%s\n    Index\t%v\n    Sequence\t%v\n",
		tx.ProposalKey.Address, tx.ProposalKey.KeyIndex, tx.ProposalKey.SequenceNumber,
	)

	if len(tx.PayloadSignatures) == 0 {
		_, _ = fmt.Fprintf(writer, "\nNo Payload Signatures\n")
	}

	if len(tx.EnvelopeSignatures) == 0 {
		_, _ = fmt.Fprintf(writer, "\nNo Envelope Signatures\n")
	}

	for i, e := range tx.PayloadSignatures {
		_, _ = fmt.Fprintf(writer, "\nPayload Signature %v:\n", i)
		_, _ = fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		_, _ = fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
		_, _ = fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
	}

	for i, e := range tx.EnvelopeSignatures {
		_, _ = fmt.Fprintf(writer, "\nEnvelope Signature %v:\n", i)
		_, _ = fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		_, _ = fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
		_, _ = fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
	}

	if tx.Script != nil {
		if len(tx.Arguments) == 0 {
			_, _ = fmt.Fprintf(writer, "\n\nArguments\tNo arguments\n")
		} else {
			_, _ = fmt.Fprintf(writer, "\n\nArguments (%d):\n", len(tx.Arguments))
			for i, argument := range tx.Arguments {
				_, _ = fmt.Fprintf(writer, "    - Argument %d: %s\n", i, argument)
			}
		}

		_, _ = fmt.Fprintf(writer, "\nCode\n\n%s\n", tx.Script)
	}

	_, _ = fmt.Fprintf(writer, "\n\n")
	_ = writer.Flush()

	prompt := promptui.Select{
		Label: promptMsg,
		Items: []string{"No", "Yes"},
	}

	_, result, _ := prompt.Run()

	_, _ = fmt.Fprintf(writer, "\r\r")
	_ = writer.Flush()

	return result == "Yes"
}

func AutocompletionPrompt() (string, string) {
	prompt := promptui.Select{
		Label: "❓ Select your shell (you can run 'echo $SHELL' to find out)",
		Items: []string{"bash", "zsh", "powershell"},
	}

	_, shell, _ := prompt.Run()
	curOs := ""

	switch shell {
	case "bash":
		prompt := promptui.Select{
			Label: "❓ Select operation system",
			Items: []string{"MacOS", "Linux"},
		}
		_, curOs, _ = prompt.Run()
	case "powershell":
		fmt.Printf(`PowerShell Installation Guide:
PS> flow config setup-completions powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> flow config setup-completions powershell > flow.ps1
# and source this file from your PowerShell profile.
`)
	}

	return shell, curOs
}

func NamePrompt() string {
	namePrompt := promptui.Prompt{
		Label: "Enter name",
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

func AccountNamePrompt(accountNames []string) string {
	namePrompt := promptui.Prompt{
		Label: "Enter an account name",
		Validate: func(s string) error {
			if slices.Contains(accountNames, s) {
				return fmt.Errorf("name already exists")
			}
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

func secureNetworkKeyPrompt() string {
	networkKeyPrompt := promptui.Prompt{
		Label: "Enter a valid host network key or leave blank",
		Validate: func(s string) error {
			if s == "" {
				return nil
			}

			return ValidateECDSAP256Pub(s)
		},
	}
	networkKey, err := networkKeyPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return networkKey
}

func addressPrompt() string {
	addressPrompt := promptui.Prompt{
		Label: "Enter address",
		Validate: func(s string) error {
			if flow.HexToAddress(s) == flow.EmptyAddress {
				return fmt.Errorf("invalid address")
			}
			return nil
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

// ShowContractDiffPrompt shows a diff between the new contract and the existing contract
// and asks the user if they wish to continue with the deployment
// returns true if the user wishes to continue with the deployment and false otherwise
func ShowContractDiffPrompt(logger output.Logger) func([]byte, []byte) bool {
	return func(newContract []byte, existingContract []byte) bool {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(newContract), string(existingContract), false)
		diffString := dmp.DiffPrettyText(diffs)
		logger.Info(diffString)

		deployPrompt := promptui.Prompt{
			Label:     "Do you wish to deploy this contract?",
			IsConfirm: true,
		}

		deploy, err := deployPrompt.Run()
		if err == promptui.ErrInterrupt {
			os.Exit(-1)
		}

		return strings.ToLower(deploy) == "y"
	}
}

type AccountData struct {
	Name     string
	Address  string
	SigAlgo  string
	HashAlgo string
	Key      string
	KeyIndex string
}

func NewAccountPrompt() *AccountData {
	var err error
	account := &AccountData{
		Name:    NamePrompt(),
		Address: addressPrompt(),
	}

	sigAlgoPrompt := promptui.Select{
		Label: "Choose signature algorithm",
		Items: []string{"ECDSA_P256", "ECDSA_secp256k1"},
	}
	_, account.SigAlgo, err = sigAlgoPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	hashAlgoPrompt := promptui.Select{
		Label: "Choose hashing algorithm",
		Items: []string{"SHA3_256", "SHA2_256"},
	}
	_, account.HashAlgo, err = hashAlgoPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	keyPrompt := promptui.Prompt{
		Label: "Enter private key",
		Validate: func(s string) error {
			_, err := crypto.DecodePrivateKeyHex(crypto.StringToSignatureAlgorithm(account.SigAlgo), s)
			return err
		},
	}
	account.Key, err = keyPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	keyIndexPrompt := promptui.Prompt{
		Label:   "Enter key index (Default: 0)",
		Default: "0",
		Validate: func(s string) error {
			v, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("invalid index, must be a number")
			}
			if v < 0 {
				return fmt.Errorf("invalid index, must be positive")
			}
			return nil
		},
	}

	account.KeyIndex, err = keyIndexPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return account
}

type ContractData struct {
	Name     string
	Source   string
	Emulator string
	Testnet  string
	Mainnet  string
}

func NewContractPrompt() *ContractData {
	contract := &ContractData{
		Name: NamePrompt(),
	}
	var err error

	sourcePrompt := promptui.Prompt{
		Label: "Enter contract file location",
		Validate: func(s string) error {
			if !config.Exists(s) {
				return fmt.Errorf("contract file doesn't exist: %s", s)
			}

			return nil
		},
	}
	contract.Source, err = sourcePrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	emulatorAliasPrompt := promptui.Prompt{
		Label: "Enter emulator alias, if exists",
		Validate: func(s string) error {
			if s != "" && flow.HexToAddress(s) == flow.EmptyAddress {
				return fmt.Errorf("invalid alias address")
			}

			return nil
		},
	}
	contract.Emulator, _ = emulatorAliasPrompt.Run()

	testnetAliasPrompt := promptui.Prompt{
		Label: "Enter testnet alias, if exists",
		Validate: func(s string) error {
			if s != "" && flow.HexToAddress(s) == flow.EmptyAddress {
				return fmt.Errorf("invalid testnet address")
			}

			return nil
		},
	}
	contract.Testnet, _ = testnetAliasPrompt.Run()

	mainnetAliasPrompt := promptui.Prompt{
		Label: "Enter mainnet alias, if exists",
		Validate: func(s string) error {
			if s != "" && flow.HexToAddress(s) == flow.EmptyAddress {
				return fmt.Errorf("invalid mainnet address")
			}

			return nil
		},
	}
	contract.Mainnet, _ = mainnetAliasPrompt.Run()

	return contract
}

func NewNetworkPrompt() map[string]string {
	networkData := make(map[string]string)
	var err error

	networkData["name"] = NamePrompt()

	hostPrompt := promptui.Prompt{
		Label: "Enter host location",
		Validate: func(s string) error {
			return nil
		},
	}
	networkData["host"], err = hostPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	networkData["key"] = secureNetworkKeyPrompt()

	return networkData
}

type DeploymentData struct {
	Network   string
	Account   string
	Contracts []string
}

func NewDeploymentPrompt(
	networks config.Networks,
	accounts config.Accounts,
	contracts config.Contracts,
) *DeploymentData {
	deploymentData := &DeploymentData{}
	var err error

	networkNames := make([]string, 0)
	for _, network := range networks {
		networkNames = append(networkNames, network.Name)
	}

	networkPrompt := promptui.Select{
		Label: "Choose network for deployment",
		Items: networkNames,
	}
	_, deploymentData.Network, err = networkPrompt.Run()
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
	_, deploymentData.Account, err = accountPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	contractNames := make([]string, 0)
	for _, contract := range contracts {
		contractNames = append(contractNames, contract.Name)
	}

	deploymentData.Contracts = make([]string, 0)

	contractName := contractPrompt(contractNames)
	deploymentData.Contracts = append(deploymentData.Contracts, contractName)
	contractNames = removeFromStringArray(contractNames, contractName)

	for addAnotherContractToDeploymentPrompt() {
		contractName := contractPrompt(contractNames)
		deploymentData.Contracts = append(deploymentData.Contracts, contractName)
		contractNames = removeFromStringArray(contractNames, contractName)

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
		Label: "Choose an account name you wish to remove",
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
		Label: "Choose deployment you wish to remove",
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
		Label: "Choose contract you wish to remove",
		Items: contractNames,
	}

	_, name, err := contractPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return name
}

func RemoveContractFromFlowJSONPrompt(contractName string) bool {
	prompt := promptui.Select{
		Label: fmt.Sprintf("Do you want to remove %s from your flow.json deployments?", contractName),
		Items: []string{"Yes", "No"},
	}
	chosen, _, _ := prompt.Run()

	return chosen == 0
}

func RemoveNetworkPrompt(networks config.Networks) string {
	networkNames := make([]string, 0)

	for _, network := range networks {
		networkNames = append(networkNames, network.Name)
	}

	networkPrompt := promptui.Select{
		Label: "Choose network you wish to remove",
		Items: networkNames,
	}

	_, name, err := networkPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return name
}

func ReportCrash() bool {
	prompt := promptui.Select{
		Label: "🙏 Please report the crash so we can improve the CLI. Do you want to report it?",
		Items: []string{"Yes, report the crash", "No"},
	}
	chosen, _, _ := prompt.Run()

	return chosen == 0
}

func CreateAccountNetworkPrompt() (string, config.Network) {
	networkMap := map[string]config.Network{
		"Emulator": config.EmulatorNetwork,
		"Testnet":  config.TestnetNetwork,
		"Mainnet":  config.MainnetNetwork,
	}

	networkPrompt := promptui.Select{
		Label: "Choose a network",
		Items: maps.Keys(networkMap),
	}

	_, selectedNetwork, err := networkPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}
	fmt.Println("")

	return selectedNetwork, networkMap[selectedNetwork]
}

func WantToUseMainnetVersionPrompt() bool {
	useMainnetVersionPrompt := promptui.Select{
		Label: "Do you wish to use Mainnet version instead? (y/n)",
		Items: []string{"Yes", "No"},
	}
	_, useMainnetVersion, err := useMainnetVersionPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return useMainnetVersion == "Yes"
}

const CancelInstall = 1

const AlreadyInstalled = 2

func InstallPrompt() int {
	prompt := promptui.Select{
		Label: "Do you wish to install it",
		Items: []string{"Yes", "No", "I've already installed it"},
	}
	index, _, err := prompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return index
}

func InstallPathPrompt(defaultPath string) string {
	prompt := promptui.Prompt{
		Label:   "Install path",
		Default: defaultPath,
		Validate: func(s string) error {
			if _, err := os.Stat(s); err == nil {
				return nil
			}

			// Attempt to create it
			var d []byte
			if err := os.WriteFile(s, d, 0644); err == nil {
				os.Remove(s) // And delete it
				return nil
			}

			return fmt.Errorf("path is invalid")
		},
	}

	install, err := prompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	return filepath.Clean(install)
}

type ScaffoldItem struct {
	Index         int
	Title         string
	Category      string
	assignedIndex int
}

func ScaffoldPrompt(logger output.Logger, scaffoldItems []ScaffoldItem) int {
	const (
		general = ""
		mobile  = "mobile"
		web     = "web"
		unity   = "unity"
	)
	outputType := map[string]string{
		general: "🔨 General Scaffolds",
		mobile:  "📱 Mobile Scaffolds",
		web:     "💻 Web Scaffolds",
		unity:   "🏀 Unity Scaffolds",
	}

	index := 0
	outputCategory := func(category string, items []ScaffoldItem) {
		logger.Info(output.Bold(output.Magenta(outputType[category])))
		for i := range items {
			if items[i].Category == category {
				index++
				logger.Info(fmt.Sprintf("   [%d] %s", index, items[i].Title))
				items[i].assignedIndex = index
			}
		}
		logger.Info("")
	}

	outputCategory(general, scaffoldItems)
	outputCategory(web, scaffoldItems)
	outputCategory(mobile, scaffoldItems)
	outputCategory(unity, scaffoldItems)

	prompt := promptui.Prompt{
		Label: "Enter the scaffold number",
		Validate: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("input must be a number")
			}

			if n < 0 && n > len(scaffoldItems) {
				return fmt.Errorf("not a valid number")
			}
			return nil
		},
	}

	input, err := prompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}
	num, _ := strconv.Atoi(input)

	for _, item := range scaffoldItems {
		if item.assignedIndex == num {
			return item.Index
		}
	}

	return 0
}
