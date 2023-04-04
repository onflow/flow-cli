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

package project

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

func Test_ProjectDeploy(t *testing.T) {
	srv, state, rw := util.TestMocks(t)

	t.Run("Fail contract errors", func(t *testing.T) {
		srv.DeployProject.Return(nil, &flowkit.ProjectDeploymentError{})
		_, err := deploy([]string{}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "failed deploying all contracts")
	})

	t.Run("Success replace standard contracts", func(t *testing.T) {
		const ft = "FungibleToken"
		const acc = "mainnet-account"
		state.Contracts().AddOrUpdate(config.Contract{
			Name:     ft,
			Location: "./ft.cdc",
		})
		_ = rw.WriteFile("./ft.cdc", []byte("test"), 0677) // mock the file
		state.Accounts().AddOrUpdate(&flowkit.Account{Name: acc, Address: flow.HexToAddress("0x01")})

		state.Deployments().AddOrUpdate(config.Deployment{
			Network:   config.MainnetNetwork.Name,
			Account:   acc,
			Contracts: []config.ContractDeployment{{Name: ft}},
		})

		state.Deployments().AddOrUpdate(config.Deployment{
			Network:   config.EmulatorNetwork.Name,
			Account:   config.DefaultEmulator.ServiceAccount,
			Contracts: []config.ContractDeployment{{Name: ft}},
		})

		err := checkForStandardContractUsageOnMainnet(state, util.NoLogger, true)
		require.NoError(t, err)

		assert.Len(t, state.Deployments().ByNetwork(config.MainnetNetwork.Name), 0)  // should remove it
		assert.Len(t, state.Deployments().ByNetwork(config.EmulatorNetwork.Name), 1) // should not remove it
		assert.NotNil(t, state.Contracts().ByName(ft).Aliases)
		assert.Equal(t, "f233dcee88fe0abe", state.Contracts().ByName(ft).Aliases.ByNetwork(config.MainnetNetwork.Name).Address.String())
	})

}
