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

package prompt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/onflow/flowkit/v2/accounts"

	"github.com/gosuri/uilive"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/exp/maps"

	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
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

	result, _ := RunSingleSelect([]string{"No", "Yes"}, promptMsg)

	_, _ = fmt.Fprintf(writer, "\r\r")
	_ = writer.Flush()

	return result == "Yes"
}

func AutocompletionPrompt() (string, string) {
	shell, _ := RunSingleSelect(
		[]string{"bash", "zsh", "powershell"},
		"❓ Select your shell (you can run 'echo $SHELL' to find out)",
	)
	curOs := ""

	switch shell {
	case "bash":
		curOs, _ = RunSingleSelect(
			[]string{"MacOS", "Linux"},
			"❓ Select operation system",
		)
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
	name, _ := RunTextInputWithValidation(
		"Enter name",
		"Type name here...",
		"",
		func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("invalid name")
			}
			return nil
		},
	)

	return name
}

func AccountNamePrompt(accountNames []string) string {
	name, _ := RunTextInputWithValidation(
		"Enter an account name",
		"my-account",
		"",
		func(s string) error {
			if slices.Contains(accountNames, s) {
				return fmt.Errorf("name already exists")
			}
			if len(s) < 1 {
				return fmt.Errorf("invalid name")
			}
			return nil
		},
	)

	return name
}

func secureNetworkKeyPrompt() string {
	networkKey, _ := RunTextInputWithValidation(
		"Enter a valid host network key or leave blank",
		"0x123456789...",
		"",
		func(s string) error {
			if s == "" {
				return nil
			}

			return util.ValidateECDSAP256Pub(s)
		},
	)

	return networkKey
}

func addressPrompt(label, errorMessage string, allowEmpty bool) string {
	placeholder := "0x1234567890abcdef"
	if allowEmpty {
		placeholder = "0x1234567890abcdef or leave blank"
	}

	address, _ := RunTextInputWithValidation(
		label,
		placeholder,
		"",
		func(s string) error {
			if allowEmpty && s == "" {
				return nil
			}
			if flow.HexToAddress(s) == flow.EmptyAddress {
				return errors.New(errorMessage)
			}
			return nil
		},
	)

	return address
}

func AddressPromptOrEmpty(label, errorMessage string) string {
	return addressPrompt(label, errorMessage, true)
}

func contractPrompt(contractNames []string) string {
	contractName, err := RunSingleSelect(
		contractNames,
		"Choose contract you wish to deploy",
	)
	if err != nil {
		os.Exit(-1)
	}

	return contractName
}

func addAnotherContractToDeploymentPrompt() bool {
	addMore, err := RunSingleSelect(
		[]string{"No", "Yes"},
		"Do you wish to add another contract for deployment?",
	)
	if err != nil {
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

		deploy, err := RunSingleSelect(
			[]string{"Yes", "No"}, 
			"Do you wish to deploy this contract?",
		)
		if err != nil {
			os.Exit(-1)
		}

		return deploy == "Yes"
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
		Address: addressPrompt("Enter address", "invalid address", false),
	}

	account.SigAlgo, err = RunSingleSelect(
		[]string{"ECDSA_P256", "ECDSA_secp256k1"},
		"Choose signature algorithm",
	)
	if err != nil {
		os.Exit(-1)
	}

	account.HashAlgo, err = RunSingleSelect(
		[]string{"SHA3_256", "SHA2_256"},
		"Choose hashing algorithm",
	)
	if err != nil {
		os.Exit(-1)
	}

	account.Key, _ = RunTextInputWithValidation(
		"Enter private key",
		"0xabcdef1234567890...",
		"",
		func(s string) error {
			_, err := crypto.DecodePrivateKeyHex(crypto.StringToSignatureAlgorithm(account.SigAlgo), s)
			return err
		},
	)

	account.KeyIndex, _ = RunTextInputWithValidation(
		"Enter key index (Default: 0)",
		"Default: 0",
		"0",
		func(s string) error {
			v, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("invalid index, must be a number")
			}
			if v < 0 {
				return fmt.Errorf("invalid index, must be positive")
			}
			return nil
		},
	)

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

	contract.Source, _ = RunTextInputWithValidation(
		"Enter contract file location",
		"cadence/contracts/HelloWorld.cdc",
		"",
		func(s string) error {
			if !config.Exists(s) {
				return fmt.Errorf("contract file doesn't exist: %s", s)
			}

			return nil
		},
	)

	contract.Emulator = AddressPromptOrEmpty("Enter emulator alias, if exists", "invalid alias address")
	contract.Testnet = AddressPromptOrEmpty("Enter testnet alias, if exists", "invalid testnet address")
	contract.Mainnet = AddressPromptOrEmpty("Enter mainnet alias, if exists", "invalid mainnet address")

	return contract
}

func NewNetworkPrompt() map[string]string {
	networkData := make(map[string]string)

	networkData["name"] = NamePrompt()

	networkData["host"], _ = RunTextInputWithValidation(
		"Enter host location",
		"http://localhost:8080",
		"",
		nil,
	)

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

	deploymentData.Network, err = RunSingleSelect(
		networkNames,
		"Choose network for deployment",
	)
	if err != nil {
		os.Exit(-1)
	}

	accountNames := make([]string, 0)
	for _, account := range accounts {
		accountNames = append(accountNames, account.Name)
	}

	deploymentData.Account, err = RunSingleSelect(
		accountNames,
		"Choose an account to deploy to",
	)
	if err != nil {
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

func removeFromStringArray(s []string, el string) []string {
	for i, v := range s {
		if v == el {
			s = slices.Delete(s, i, i+1)
			break
		}
	}

	return s
}

// AddContractToDeploymentPrompt prompts a user to select an account to deploy a given contract on a given network
func AddContractToDeploymentPrompt(networkName string, accounts accounts.Accounts, contractName string) *DeploymentData {
	deploymentData := &DeploymentData{
		Network:   networkName,
		Contracts: []string{contractName},
	}
	var err error

	accountNames := make([]string, 0)
	for _, account := range accounts {
		accountNames = append(accountNames, account.Name)
	}

	// Add a "none" option to the list of accounts
	accountNames = append(accountNames, "none")

	selectedAccount, err := RunSingleSelect(
		accountNames,
		fmt.Sprintf("Choose an account to deploy %s to on %s (or 'none' to skip)", contractName, networkName),
	)
	if err != nil {
		os.Exit(-1)
	}

	// Handle the "none" selection
	if selectedAccount == "none" {
		return nil
	}

	deploymentData.Account = selectedAccount

	return deploymentData
}

func RemoveAccountPrompt(accounts config.Accounts) string {
	accountNames := make([]string, 0)

	for _, account := range accounts {
		accountNames = append(accountNames, account.Name)
	}

	name, err := RunSingleSelect(
		accountNames,
		"Choose an account name you wish to remove",
	)
	if err != nil {
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

	selectedDeployment, err := RunSingleSelect(
		deploymentNames,
		"Choose deployment you wish to remove",
	)
	if err != nil {
		os.Exit(-1)
	}

	// Find the index of the selected deployment
	var selectedIndex int
	for i, deploymentName := range deploymentNames {
		if deploymentName == selectedDeployment {
			selectedIndex = i
			break
		}
	}

	return deployments[selectedIndex].Account, deployments[selectedIndex].Network
}

func RemoveContractPrompt(contracts config.Contracts) string {
	contractNames := make([]string, 0)

	for _, contract := range contracts {
		contractNames = append(contractNames, contract.Name)
	}

	name, err := RunSingleSelect(
		contractNames,
		"Choose contract you wish to remove",
	)
	if err != nil {
		os.Exit(-1)
	}

	return name
}

func RemoveContractFromFlowJSONPrompt(contractName string) bool {
	chosen, _ := RunSingleSelect(
		[]string{"Yes", "No"},
		fmt.Sprintf("Do you want to remove %s from your flow.json deployments?", contractName),
	)

	return chosen == "Yes"
}

func RemoveNetworkPrompt(networks config.Networks) string {
	networkNames := make([]string, 0)

	for _, network := range networks {
		networkNames = append(networkNames, network.Name)
	}

	name, err := RunSingleSelect(
		networkNames,
		"Choose network you wish to remove",
	)
	if err != nil {
		os.Exit(-1)
	}

	return name
}

func ReportCrash() bool {
	chosen, _ := RunSingleSelect(
		[]string{"Yes, report the crash", "No"},
		"🙏 Please report the crash so we can improve the CLI. Do you want to report it?",
	)

	return chosen == "Yes, report the crash"
}

func CreateAccountNetworkPrompt() (string, config.Network) {
	networkMap := map[string]config.Network{
		"Emulator": config.EmulatorNetwork,
		"Testnet":  config.TestnetNetwork,
		"Mainnet":  config.MainnetNetwork,
	}

	selectedNetwork, err := RunSingleSelect(
		maps.Keys(networkMap),
		"Choose a network",
	)
	if err != nil {
		os.Exit(-1)
	}
	fmt.Println("")

	return selectedNetwork, networkMap[selectedNetwork]
}

func WantToUseMainnetVersionPrompt() bool {
	useMainnetVersion, err := RunSingleSelect(
		[]string{"Yes", "No"},
		"Do you wish to use Mainnet version instead? (y/n)",
	)
	if err != nil {
		os.Exit(-1)
	}

	return useMainnetVersion == "Yes"
}

const CancelInstall = 1

const AlreadyInstalled = 2

func InstallPrompt() int {
	options := []string{"Yes", "No", "I've already installed it"}
	selection, err := RunSingleSelect(
		options,
		"Do you wish to install it",
	)
	if err != nil {
		os.Exit(-1)
	}

	// Find the index of the selection
	for i, option := range options {
		if option == selection {
			return i
		}
	}

	return 0 // fallback to "Yes"
}

func InstallPathPrompt(defaultPath string) string {
	install, _ := RunTextInputWithValidation(
		"Install path",
		defaultPath,
		defaultPath,
		func(s string) error {
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
	)

	return filepath.Clean(install)
}

func GenericBoolPrompt(msg string) bool {
	result, _ := RunSingleSelect([]string{"Yes", "No"}, msg)

	return result == "Yes"
}

func GenericSelect(items []string, message string) string {
	result, _ := RunSingleSelect(items, message)

	return result
}
