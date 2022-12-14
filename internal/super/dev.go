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
	"fmt"
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
	state *flowkit.State,
) (command.Result, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// todo get from flag
	network := config.DefaultEmulatorNetwork().Name

	// todo dev work if not run on top root directory - at least have a warning
	// todo handle emulator running as part of this service or part of existing running emulator

	service, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}

	services.SetLogger(output.NewStdoutLogger(output.NoneLog))

	proj := newProjectFiles(dir)

	deployments, err := proj.deployments()
	if err != nil {
		return nil, err
	}

	err = startup(deployments, network, service, services, state, readerWriter)
	if err != nil {
		return nil, err
	}

	err = deploy(network, services)
	if err != nil {
		return nil, err
	}

	accountChanges, contractChanges, err := proj.watch()
	if err != nil {
		return nil, err
	}

	for {
		select {
		case account := <-accountChanges:
			var err error
			if account.status == created {
				err = addAccount(account.name, service, services, state)
			}
			if account.status == removed {
				err = state.Accounts().Remove(account.name)
			}
			if err != nil {
				return nil, err
			}
			err = state.SaveDefault()
			if err != nil {
				return nil, err
			}
		case contract := <-contractChanges:
			if contract.status == created {
				_, err := addContract(contract.path, state, readerWriter)
				if err != nil {
					return nil, err
				}
			}
			if contract.status == removed {
				err := removeContract(contract.path, contract.account, network, state, readerWriter)
				if err != nil {
					return nil, err
				}
			}
			err := deploy(network, services)
			if err != nil {
				return nil, err
			}

			err = state.SaveDefault()
			if err != nil {
				return nil, err
			}
		}
	}
}

func deploy(network string, services *services.Services) error {
	deployed, err := services.Project.Deploy(network, true)
	if err != nil {
		return err
	}

	for _, d := range deployed {
		fmt.Printf("deployed [%s] on account [%s]\n", d.Name, d.AccountName)
	}
	return nil
}

func startup(
	deployments map[string][]string,
	network string,
	service *flowkit.Account,
	services *services.Services,
	state *flowkit.State,
	readerWriter flowkit.ReaderWriter,
) error {
	cleanState(state)

	for accName, contracts := range deployments {
		if accName == "" { // default to emulator account
			accName = config.DefaultEmulatorServiceAccountName
		}

		err := addAccount(accName, service, services, state)
		if err != nil {
			return err
		}

		contractDeployments := make([]config.ContractDeployment, len(contracts))
		for i, path := range contracts {
			contract, err := addContract(path, state, readerWriter)
			if err != nil {
				return err
			}
			contractDeployments[i] = config.ContractDeployment{Name: contract.Name}
		}

		state.Deployments().AddOrUpdate(config.Deployment{
			Network:   network,
			Account:   accName,
			Contracts: contractDeployments,
		})
	}

	return state.SaveDefault()
}

// cleanState of existing contracts, deployments and non-service accounts as we will build it again.
func cleanState(state *flowkit.State) {
	for _, c := range *state.Contracts() {
		_ = state.Contracts().Remove(c.Name)
	}
	for _, d := range *state.Deployments() {
		_ = state.Deployments().Remove(d.Account, d.Network)
	}
	// clean out non-service accounts
	accs := make([]flowkit.Account, len(*state.Accounts()))
	copy(accs, *state.Accounts()) // we need to make a copy otherwise when we remove order shifts
	for _, a := range accs {
		if a.Name() == config.DefaultEmulatorServiceAccountName {
			continue
		}
		_ = state.Accounts().Remove(a.Name())
	}
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

func contractName(path string, readerWriter flowkit.ReaderWriter) (string, error) {
	// todo add warning if name of the file is not matching the name of the contract
	content, err := readerWriter.ReadFile(path)
	if err != nil {
		return "", err
	}

	program, err := project.NewProgram(flowkit.NewScript(content, nil, path))
	if err != nil {
		return "", err
	}

	name, err := program.Name()
	if err != nil {
		return "", err
	}

	return name, nil
}

func addContract(
	path string,
	state *flowkit.State,
	readerWriter flowkit.ReaderWriter,
) (*config.Contract, error) {
	name, err := contractName(path, readerWriter)
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

func removeContract(
	path string,
	accountName string,
	network string,
	state *flowkit.State,
	readerWriter flowkit.ReaderWriter,
) error {
	name, err := contractName(path, readerWriter)
	if err != nil {
		return err
	}

	if accountName == "" {
		accountName = config.DefaultEmulatorServiceAccountName
	}

	state.Deployments().RemoveContract(accountName, network, name)
	return nil
}
