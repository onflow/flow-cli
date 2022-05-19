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
	"bytes"
	"fmt"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type FlagsInit struct {
	ServicePrivateKey  string `flag:"service-private-key" info:"Service account private key"`
	ServiceKeySigAlgo  string `default:"ECDSA_P256" flag:"service-sig-algo" info:"Service account key signature algorithm"`
	ServiceKeyHashAlgo string `default:"SHA3_256" flag:"service-hash-algo" info:"Service account key hash algorithm"`
	Reset              bool   `default:"false" flag:"reset" info:"Reset configuration file"`
	Global             bool   `default:"false" flag:"global" info:"Initialize global user configuration"`
}

var InitFlag = FlagsInit{}

var InitCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration",
	},
	Flags: &InitFlag,
	Run:   Initialise,
}

func Initialise(
	_ []string,
	readerWriter flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {

	sigAlgo := crypto.StringToSignatureAlgorithm(InitFlag.ServiceKeySigAlgo)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("invalid signature algorithm: %s", InitFlag.ServiceKeySigAlgo)
	}

	hashAlgo := crypto.StringToHashAlgorithm(InitFlag.ServiceKeyHashAlgo)
	if hashAlgo == crypto.UnknownHashAlgorithm {
		return nil, fmt.Errorf("invalid hash algorithm: %s", InitFlag.ServiceKeyHashAlgo)
	}

	var privateKey crypto.PrivateKey
	if InitFlag.ServicePrivateKey != "" {
		var err error
		privateKey, err = crypto.DecodePrivateKeyHex(sigAlgo, InitFlag.ServicePrivateKey)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}
	}

	s, err := services.Project.Init(
		readerWriter,
		InitFlag.Reset,
		InitFlag.Global,
		sigAlgo,
		hashAlgo,
		privateKey,
	)
	if err != nil {
		return nil, err
	}

	return &InitResult{State: s}, nil
}

type InitResult struct {
	*flowkit.State
}

func (r *InitResult) JSON() interface{} {
	return r
}

func (r *InitResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)
	account, _ := r.State.EmulatorServiceAccount()

	_, _ = fmt.Fprintf(writer, "Configuration initialized\n")
	_, _ = fmt.Fprintf(writer, "Service account: %s\n\n", output.Bold("0x"+account.Address().String()))
	_, _ = fmt.Fprintf(writer,
		"Start emulator by running: %s \nReset configuration using: %s\n",
		output.Bold("'flow emulator'"),
		output.Bold("'flow init --reset'"),
	)

	_ = writer.Flush()
	return b.String()
}

func (r *InitResult) Oneliner() string {
	return ""
}
