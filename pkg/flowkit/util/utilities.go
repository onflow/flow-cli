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
	"path"
	"strings"
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

// GetAddressNetwork returns the chain ID for an address.
func GetAddressNetwork(address flow.Address) (flow.ChainID, error) {
	networks := []flow.ChainID{
		flow.Mainnet,
		flow.Testnet,
		flow.Emulator,
		flow.Sandboxnet,
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
		address.IsValid(flow.Emulator) ||
		address.IsValid(flow.Sandboxnet)
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

func AbsolutePath(basePath, filePath string) string {
	if path.IsAbs(filePath) {
		return filePath
	}

	return path.Join(path.Dir(basePath), filePath)
}

type ScriptQuery struct {
	ID     flow.Identifier
	Height uint64
}
