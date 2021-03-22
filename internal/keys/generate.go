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

package keys

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsGenerate struct {
	Seed       string `flag:"seed" info:"Deterministic seed phrase"`
	KeySigAlgo string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm"`
}

type cmdGenerate struct {
	cmd   *cobra.Command
	flags flagsGenerate
}

// NewGenerateCmd return new command
func NewGenerateCmd() command.Command {
	return &cmdGenerate{
		cmd: &cobra.Command{
			Use:   "generate",
			Short: "Generate a new key-pair",
		},
	}
}

// Run command
func (a *cmdGenerate) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {
	keys, err := services.Keys.Generate(a.flags.Seed, a.flags.KeySigAlgo)
	if err != nil {
		return nil, err
	}

	pubKey := keys.PublicKey()
	return &KeyResult{privateKey: keys, publicKey: &pubKey}, nil
}

// GetFlags get command flags
func (a *cmdGenerate) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

// GetCmd gets command
func (a *cmdGenerate) GetCmd() *cobra.Command {
	return a.cmd
}
