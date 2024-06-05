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
	"os"

	"github.com/onflow/flowkit/v2/accounts"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
)

// InitConfigParameters holds all necessary parameters for initializing the configuration.
type InitConfigParameters struct {
	ServicePrivateKey  string
	ServiceKeySigAlgo  string
	ServiceKeyHashAlgo string
	Reset              bool
	Global             bool
	TargetDirectory    string
}

// InitializeConfiguration creates the Flow configuration json file based on the provided parameters.
func InitializeConfiguration(params InitConfigParameters, readerWriter flowkit.ReaderWriter) (*flowkit.State, error) {
	var path string
	if params.TargetDirectory != "" {
		path = fmt.Sprintf("%s/flow.json", params.TargetDirectory)

		// Create the directory if it doesn't exist
		err := readerWriter.MkdirAll(params.TargetDirectory, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("failed to create target directory: %w", err)
		}
	} else {
		// Otherwise, choose between the default and global paths
		if params.Global {
			path = config.GlobalPath()
		} else {
			path = config.DefaultPath
		}
	}

	sigAlgo := crypto.StringToSignatureAlgorithm(params.ServiceKeySigAlgo)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("invalid signature algorithm: %s", params.ServiceKeySigAlgo)
	}

	hashAlgo := crypto.StringToHashAlgorithm(params.ServiceKeyHashAlgo)
	if hashAlgo == crypto.UnknownHashAlgorithm {
		return nil, fmt.Errorf("invalid hash algorithm: %s", params.ServiceKeyHashAlgo)
	}

	state, err := flowkit.Init(readerWriter)
	if err != nil {
		return nil, err
	}

	emulatorAccount, err := accounts.NewEmulatorAccount(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256, params.TargetDirectory)
	if err != nil {
		return nil, err
	}

	state.Accounts().AddOrUpdate(emulatorAccount)

	if params.ServicePrivateKey != "" {
		privateKey, err := crypto.DecodePrivateKeyHex(sigAlgo, params.ServicePrivateKey)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}

		state.SetEmulatorKey(privateKey)
	}

	if config.Exists(path) && !params.Reset {
		return nil, fmt.Errorf(
			"configuration already exists at: %s, if you want to reset configuration use the reset flag",
			path,
		)
	}

	return state, nil
}
