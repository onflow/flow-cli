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

package output

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gosuri/uilive"
	"github.com/manifoldco/promptui"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

func ApproveTransactionForSigningPrompt(transaction *flowkit.Transaction) bool {
	return ApproveTransactionPrompt(transaction, "⚠️  Do you want to SIGN this transaction?")
}

func ApproveTransactionForBuildingPrompt(transaction *flowkit.Transaction) bool {
	return ApproveTransactionPrompt(transaction, "⚠️  Do you want to BUILD this transaction?")
}

func ApproveTransactionForSendingPrompt(transaction *flowkit.Transaction) bool {
	return ApproveTransactionPrompt(transaction, "⚠️  Do you want to SEND this transaction?")
}

func ApproveTransactionPrompt(transaction *flowkit.Transaction, promptMsg string) bool {
	writer := uilive.New()
	tx := transaction.FlowTransaction()

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

func AccountNamePrompt(accounts *flowkit.Accounts) string {
	existingNames := make([]string, len(*accounts))
	for i, a := range *accounts {
		existingNames[i] = a.Name()
	}

	namePrompt := promptui.Prompt{
		Label: "Enter an account name",
		Validate: func(s string) error {
			if slices.Contains(existingNames, s) {
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

			return util.ValidateECDSAP256Pub(s)
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

	accountData["name"] = NamePrompt()
	accountData["address"] = addressPrompt()

	sigAlgoPrompt := promptui.Select{
		Label: "Choose signature algorithm",
		Items: []string{"ECDSA_P256", "ECDSA_secp256k1"},
	}
	_, accountData["sigAlgo"], err = sigAlgoPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	hashAlgoPrompt := promptui.Select{
		Label: "Choose hashing algorithm",
		Items: []string{"SHA3_256", "SHA2_256"},
	}
	_, accountData["hashAlgo"], err = hashAlgoPrompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	keyPrompt := promptui.Prompt{
		Label: "Enter private key",
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
		Label:   "Enter key index (Default: 0)",
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

	contractData["name"] = NamePrompt()

	sourcePrompt := promptui.Prompt{
		Label: "Enter contract file location",
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
		"Emulator": config.DefaultEmulatorNetwork(),
		"Testnet":  config.DefaultTestnetNetwork(),
		"Mainnet":  config.DefaultMainnetNetwork(),
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

func WantToContinue() bool {
	prompt := promptui.Prompt{
		Label:       "Do you want to continue",
		IsConfirm:   true,
		HideEntered: true,
	}
	selected, err := prompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}

	fmt.Print("\033[1A\033[K")
	return strings.ToLower(selected) == "y"
}

func ConfirmOpenBrowser() {
	prompt := promptui.Prompt{
		Label:       "Press <ENTER> to open in your browser...",
		HideEntered: true,
	}
	_, err := prompt.Run()
	if err == promptui.ErrInterrupt {
		os.Exit(-1)
	}
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

	return path.Clean(install)
}

func ScaffoldPrompt(logger Logger, availableScaffolds map[string][]string) int {
	index := 0
	for t, items := range availableScaffolds {
		logger.Info(t)
		for _, item := range items {
			index++
			logger.Info(fmt.Sprintf("  [%d] %s", index, item))
		}
		logger.Info("") // new line
	}

	prompt := promptui.Prompt{
		Label:   "Enter the scaffold number",
		Default: "1",
		Validate: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("input must be a number")
			}

			if n > len(availableScaffolds) {
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
	return num
}
