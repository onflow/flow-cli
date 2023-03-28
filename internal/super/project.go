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
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/pkg/errors"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	flowkitProject "github.com/onflow/flow-cli/pkg/flowkit/project"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

var emulator = config.DefaultEmulatorNetwork().Name

const defaultAccount = "default"

func newProject(
	serviceAccount flowkit.Account,
	services *services.Services,
	state *flowkit.State,
	readerWriter flowkit.ReaderWriter,
	files *projectFiles,
) (*project, error) {
	proj := &project{
		service:        &serviceAccount,
		services:       services,
		state:          state,
		readerWriter:   readerWriter,
		projectFiles:   files,
		pathNameLookup: make(map[string]string),
	}

	if err := proj.projectFiles.exist(); err != nil {
		return nil, err
	}

	return proj, nil
}

type project struct {
	service        *flowkit.Account
	services       *services.Services
	state          *flowkit.State
	readerWriter   flowkit.ReaderWriter
	projectFiles   *projectFiles
	pathNameLookup map[string]string
}

// startup cleans the state and then rebuilds it from the current folder state.
func (p *project) startup() error {
	deployments, err := p.projectFiles.deployments()
	if err != nil {
		return err
	}

	p.cleanState()
	err = p.addAccount(defaultAccount)
	if err != nil {
		return err
	}

	for accName, contracts := range deployments {
		if accName == "" { // default to emulator account
			accName = defaultAccount
		}

		err := p.addAccount(accName)
		if err != nil {
			return err
		}

		p.state.Deployments().AddOrUpdate(config.Deployment{
			Network: emulator,
			Account: accName,
		})
		for _, path := range contracts {
			err := p.addContract(path, accName)
			if err != nil {
				return err
			}
		}
	}

	p.deploy()

	return p.state.SaveDefault()
}

// deploys all the contracts found in the state configuration.
func (p *project) deploy() {
	deployed, err := p.services.Project.Deploy(emulator, services.UpdateExisting(true))
	printDeployment(deployed, err, p.pathNameLookup)
}

// cleanState of existing contracts, deployments and non-service accounts as we will build it again.
func (p *project) cleanState() {
	contracts := make(config.Contracts, len(*p.state.Contracts()))
	copy(contracts, *p.state.Contracts()) // we need to make a copy otherwise when we remove order shifts
	for _, c := range contracts {
		if c.IsAliased() {
			continue // don't remove aliased contracts
		}

		_ = p.state.Contracts().Remove(c.Name)
	}

	for _, d := range p.state.Deployments().ByNetwork(emulator) {
		_ = p.state.Deployments().Remove(d.Account, emulator)
	}

	accs := make([]flowkit.Account, len(*p.state.Accounts()))
	copy(accs, *p.state.Accounts()) // we need to make a copy otherwise when we remove order shifts
	for _, a := range accs {
		chain, err := util.GetAddressNetwork(a.Address())
		if err != nil || chain != flow.Emulator {
			continue // don't remove non-emulator accounts
		}

		if a.Name() == config.DefaultEmulatorServiceAccountName {
			continue
		}
		_ = p.state.Accounts().Remove(a.Name())
	}
}

// watch project files and update the state accordingly.
func (p *project) watch() error {
	accountChanges, contractChanges, err := p.projectFiles.watch()
	if err != nil {
		return errors.Wrap(err, "error watching files")
	}

	for {
		select {
		case account := <-accountChanges:
			if account.status == created {
				err = p.addAccount(account.name)
			}
			if account.status == removed {
				err = p.removeAccount(account.name)
			}
			if err != nil {
				return errors.Wrap(err, "failed updating accounts")
			}
		case contract := <-contractChanges:
			if contract.account == "" {
				contract.account = defaultAccount
			}

			switch contract.status {
			case created:
				_ = p.addContract(contract.path, contract.account)
			case changed:
				_ = p.addContract(contract.path, contract.account)
			case renamed:
				p.renameContract(contract.oldPath, contract.path)
			case removed:
				err = p.removeContract(contract.path, contract.account) // todo what if contract got broken and then we want to delete it
				if err != nil {
					return err
				}
			}

			p.deploy()
		}

		err = p.state.SaveDefault()
		if err != nil {
			return errors.Wrap(err, "failed saving configuration")
		}
	}
}

// addAccount to the state and create it on the network.
func (p *project) addAccount(name string) error {
	pkey, err := p.services.Keys.Generate("", crypto.ECDSA_P256)
	if err != nil {
		return err
	}

	// create the account on the network and set the address
	flowAcc, err := p.services.Accounts.Create(
		p.service,
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

	p.state.Accounts().AddOrUpdate(account)
	p.state.Deployments().AddOrUpdate(config.Deployment{ // init empty deployment
		Network: emulator,
		Account: name,
	})
	return nil
}

func (p *project) removeAccount(name string) error {
	_ = p.state.Deployments().Remove(name, emulator)
	return p.state.Accounts().Remove(name)
}

// contractName extracts contract name from the source code.
func (p *project) contractName(path string) (string, error) {
	if name, ok := p.pathNameLookup[path]; ok {
		return name, nil
	}

	// todo add warning if name of the file is not matching the name of the contract
	content, err := p.readerWriter.ReadFile(path)
	if err != nil {
		return "", errors.Wrap(err, "could not load contract to get the name")
	}

	program, err := flowkitProject.NewProgram(flowkit.NewScript(content, nil, path))
	if err != nil {
		return "", err
	}

	name, err := program.Name()
	if err != nil {
		return "", err
	}

	p.pathNameLookup[path] = name

	return name, nil
}

// addContract to the state configuration as a contract and as a deployment.
func (p *project) addContract(
	path string,
	account string,
) error {
	name, err := p.contractName(path)
	if err != nil {
		return err
	}

	contract := config.Contract{
		Name:     name,
		Location: path,
	}

	existing := p.state.Contracts().ByName(name)
	if existing != nil { // make sure alises are persisted even if location changes
		contract.Aliases = existing.Aliases
	}

	if contract.Aliases.ByNetwork(emulator) == nil { // only add if not existing emulator alias
		deployment := config.ContractDeployment{
			Name: contract.Name,
		}
		p.state.Deployments().AddContract(account, emulator, deployment)
	}

	p.state.Contracts().AddOrUpdate(contract)
	return nil
}

// removeContract from state configuration.
func (p *project) removeContract(
	path string,
	accountName string,
) error {
	name, err := p.contractName(path)
	if err != nil {
		return errors.Wrap(err, "failed to remove contract")
	}

	if accountName == "" {
		accountName = defaultAccount
	}

	if len(p.state.Deployments().ByAccountAndNetwork(accountName, emulator)) > 0 {
		p.state.Deployments().RemoveContract(accountName, emulator, name) // we might delete account first
	}

	return nil
}

// renameContract and update the location in the state
func (p *project) renameContract(oldLocation string, newLocation string) {
	for _, c := range *p.state.Contracts() {
		if c.Location == oldLocation {
			c.Location = newLocation
			p.state.Contracts().AddOrUpdate(c)
		}
	}
}
