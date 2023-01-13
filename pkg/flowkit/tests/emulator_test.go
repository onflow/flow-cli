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

package tests

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var emulator = config.DefaultEmulatorNetwork().Name

func initEmulator(t *testing.T) (gateway.Gateway, *flowkit.State, *services.Services, afero.Fs) {
	readerWriter, mockFs := ReaderWriter()

	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	serviceAccount, err := state.EmulatorServiceAccount()
	require.NoError(t, err)

	gw := gateway.NewEmulatorGateway(serviceAccount)
	logger := output.NewStdoutLogger(output.NoneLog)
	srv := services.NewServices(gw, state, logger)

	return gw, state, srv, mockFs
}

func Test_Emulator_ProjectDeploy(t *testing.T) {
	_, state, srv, mockFs := initEmulator(t)

	state.Contracts().AddOrUpdate(ContractA.Name, config.Contract{
		Name:    ContractA.Name,
		Source:  ContractA.Filename,
		Network: emulator,
	})

	state.Contracts().AddOrUpdate(ContractB.Name, config.Contract{
		Name:    ContractB.Name,
		Source:  ContractB.Filename,
		Network: emulator,
	})

	state.Contracts().AddOrUpdate(ContractC.Name, config.Contract{
		Name:    ContractC.Name,
		Source:  ContractC.Filename,
		Network: emulator,
	})

	serviceAcc, err := state.EmulatorServiceAccount()
	require.NoError(t, err)

	initArg, _ := cadence.NewString("foo")
	state.Deployments().AddOrUpdate(config.Deployment{
		Network: emulator,
		Account: serviceAcc.Name(),
		Contracts: []config.ContractDeployment{
			{Name: ContractA.Name},
			{Name: ContractB.Name},
			{Name: ContractC.Name, Args: []cadence.Value{initArg}},
		},
	})

	contracts, err := srv.Project.Deploy(emulator, true)
	assert.NoError(t, err)
	assert.Len(t, contracts, 3)
	assert.Equal(t, ContractA.Name, contracts[0].Name())
	assert.Equal(t, ContractB.Name, contracts[1].Name())
	assert.Equal(t, ContractC.Name, contracts[2].Name())

	// make a change
	ContractA.Source = []byte(`pub contract ContractA { init() {} }`)
	_ = afero.WriteFile(mockFs, ContractA.Filename, ContractA.Source, 0644)

	contracts, err = srv.Project.Deploy(emulator, true)
	assert.NoError(t, err)
	assert.Len(t, contracts, 3)
	assert.Equal(t, ContractA.Name, contracts[0].Name())
	assert.Equal(t, ContractB.Name, contracts[1].Name())
	assert.Equal(t, ContractC.Name, contracts[2].Name())
}
