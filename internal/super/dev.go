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

package super

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"os"

	"github.com/spf13/cobra"
)

type flagsDev struct{}

var devFlags = flagsDev{}

var DevCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "dev",
		Short:   "Monitor your project during development", // todo improve
		Args:    cobra.ExactArgs(0),
		Example: "flow dev",
	},
	Flags: &devFlags,
	RunS:  dev,
}

func dev(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
	s *flowkit.State,
) (command.Result, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// todo dev work if not run on top root directory - at least have a warning
	// todo handle emulator running as part of this service or part of existing running emulator

	service, err := s.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}

	logger := output.NewStdoutLogger(output.NoneLog)
	services.SetLogger(logger)

	// todo maybe not optimal, test how it performs otherwise we will just keep the state and update changes, especially the account section is problematic, creating accounts everytime might be time consuming
	// create new state everytime, just keep the service key the same
	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		return nil, err
	}

	serviceKey, err := service.Key().PrivateKey()
	if err != nil {
		return nil, err
	}
	state.SetEmulatorKey(*serviceKey)

	proj := newProjectFiles(dir)

	deployments, err := proj.Deployments()
	if err != nil {
		return nil, err
	}

	for accName, contracts := range deployments {
		if accName == "" {
			accName = config.DefaultEmulatorServiceAccountName
		}

		err := addAccount(accName, service, services, state)
		if err != nil {
			return nil, err
		}

		contractDeployments := make([]config.ContractDeployment, len(contracts))
		for i, path := range contracts {
			contract, err := addContract(path, state, readerWriter)
			if err != nil {
				return nil, err
			}
			contractDeployments[i] = config.ContractDeployment{Name: contract.Name}
		}

		state.Deployments().AddOrUpdate(config.Deployment{
			Network:   config.DefaultEmulatorNetwork().Name, // todo take in from flag
			Account:   accName,
			Contracts: contractDeployments,
		})
	}

	err = state.SaveDefault()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func addAccount(name string, service *flowkit.Account, services *services.Services, state *flowkit.State) error {
	pkey, err := services.Keys.Generate("", crypto.ECDSA_P256)
	if err != nil {
		return err
	}

	// create the account on the network and set the address
	flowAcc, err := services.Accounts.Create(
		service,
		[]crypto.PublicKey{pkey.PublicKey()},
		[]int{flow.AccountKeyWeightThreshold},
		[]crypto.SignatureAlgorithm{crypto.ECDSA_P256},
		[]crypto.HashAlgorithm{crypto.SHA3_256},
		nil,
	)
	if err != nil {
		return err
	}

	account := flowkit.NewAccount(name)
	account.SetAddress(flowAcc.Address)
	account.SetKey(flowkit.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, pkey))

	state.Accounts().AddOrUpdate(account)
	return nil
}

func addContract(path string, state *flowkit.State, readerWriter flowkit.ReaderWriter) (*config.Contract, error) {
	content, err := readerWriter.ReadFile(path)
	if err != nil {
		return nil, err
	}

	program, err := project.NewProgram(flowkit.NewScript(content, nil, path))
	if err != nil {
		return nil, err
	}

	name, err := program.Name()
	if err != nil {
		return nil, err
	}

	contract := config.Contract{
		Name:     name,
		Location: path,
	}
	state.Contracts().AddOrUpdate(name, contract)
	return &contract, nil
}
