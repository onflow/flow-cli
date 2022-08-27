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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

const TestnetFaucetHost = "https://testnet-faucet.onflow.org/"
const FlowPortUrl = "https://port.onflow.org/transaction?hash=a0a78aa7821144efd5ebb974bb52ba04609ce76c3863af9d45348db93937cf98&showcode=false&consent=true&pk="
const FlowCLIConfigFile = "config.json"

// ConvertSigAndHashAlgo parses and validates a signature and hash algorithm pair.
func ConvertSigAndHashAlgo(
	signatureAlgorithm string,
	hashingAlgorithm string,
) (crypto.SignatureAlgorithm, crypto.HashAlgorithm, error) {
	sigAlgo := crypto.StringToSignatureAlgorithm(signatureAlgorithm)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return crypto.UnknownSignatureAlgorithm,
			crypto.UnknownHashAlgorithm,
			fmt.Errorf("failed to determine signature algorithm from %s", sigAlgo)
	}

	hashAlgo := crypto.StringToHashAlgorithm(hashingAlgorithm)
	if hashAlgo == crypto.UnknownHashAlgorithm {
		return crypto.UnknownSignatureAlgorithm,
			crypto.UnknownHashAlgorithm,
			fmt.Errorf("failed to determine hash algorithm from %s", hashAlgo)
	}

	return sigAlgo, hashAlgo, nil
}

// ContainsString returns true if the slice contains the given string.
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
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

func ParseAddress(value string) (flow.Address, bool) {
	address := flow.HexToAddress(value)

	// valid on any chain
	return address, address.IsValid(flow.Mainnet) ||
		address.IsValid(flow.Testnet) ||
		address.IsValid(flow.Emulator)
}

func RemoveFromStringArray(s []string, el string) []string {
	for i, v := range s {
		if v == el {
			s = append(s[:i], s[i+1:]...)
			break
		}
	}

	return s
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

func OpenBrowserWindow(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return fmt.Errorf("could not open a browser window, please navigate to %s manually: %w", url, err)
	}
	return nil
}

func TestnetFaucetURL(publicKey string, sigAlgo crypto.SignatureAlgorithm) string {

	link := fmt.Sprintf("%s?key=%s", TestnetFaucetHost, strings.TrimPrefix(publicKey, "0x"))
	if sigAlgo != crypto.ECDSA_P256 {
		link = fmt.Sprintf("%s&sig-algo=%s", link, sigAlgo)
	}

	return link
}

func MainnetFlowPortURL(publicKey string) string {
	return fmt.Sprintf("%s%s", FlowPortUrl, strings.TrimPrefix(publicKey, "0x"))
}

type ReaderWriter interface {
	ReadFile(source string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// AddToGitIgnore adds a new line to the .gitignore if one doesn't exist it creates it.
func AddToGitIgnore(filename string, loader ReaderWriter) error {
	currentWd, err := os.Getwd()
	if err != nil {
		return err
	}
	gitIgnorePath := path.Join(currentWd, ".gitignore")
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

type jsonConfig struct {
	Metrics bool `json:"metrics"`
}

// AddToConfig adds a new line to user's config dir stating the user's preference for command tracking
func AddToConfig(loader ReaderWriter, metricsEnabled bool) (string, error) {
	homeConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	cliDir := filepath.Join(homeConfigDir, FLOW_CLI)
	if _, err := os.Stat(cliDir); os.IsNotExist(err) {
		os.Mkdir(cliDir, 0755)
	}
	cliConfigPath := filepath.Join(cliDir, FlowCLIConfigFile)
	filePermissions := os.FileMode(0644)

	fileStat, err := os.Stat(cliConfigPath)
	if !os.IsNotExist(err) { // if config.json exists
		filePermissions = fileStat.Mode().Perm()
	}
	jsonConfig := jsonConfig{
		Metrics: metricsEnabled,
	}
	data, err := json.MarshalIndent(jsonConfig, "", " ")

	return cliConfigPath, loader.WriteFile(cliConfigPath, data, filePermissions)
}

// CheckMetricsEnabled checks if a user is opted in to have command tracked with Mixpanel
// Returns true if user has not opted out before
func CheckMetricsEnabled(loader ReaderWriter) (bool, error) {
	var jsonConf jsonConfig
	homeConfigDir, err := os.UserConfigDir()
	if err != nil {
		return false, err
	}
	cliDir := filepath.Join(homeConfigDir, FLOW_CLI)
	cliConfigPath := filepath.Join(cliDir, FlowCLIConfigFile)

	_, err = os.Stat(cliConfigPath)
	if os.IsNotExist(err) { // if config.json exists, return metrics enabled by default
		return true, nil
	}

	raw, err := loader.ReadFile(cliConfigPath)
	err = json.Unmarshal(raw, &jsonConf)
	if err != nil {
		return false, err
	}

	return jsonConf.Metrics, nil
}
