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

package dependencymanager

import (
	"fmt"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/gateway"
	"github.com/onflow/flow-cli/flowkit/gateway/mocks"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/flowkit/tests"
	"github.com/onflow/flow-cli/internal/util"
)

func TestDependencyInstallerInstall(t *testing.T) {

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address

	dep := config.Dependency{
		Name: "Hello",
		RemoteSource: config.RemoteSource{
			NetworkName:  "emulator",
			Address:      serviceAddress,
			ContractName: "Hello",
		},
	}

	state.Dependencies().AddOrUpdate(dep)

	t.Run("Success", func(t *testing.T) {
		gw := mocks.DefaultMockGateway()

		gw.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(0).(flow.Address)
			assert.Equal(t, addr.String(), serviceAcc.Address.String())
			racc := tests.NewAccountWithAddress(addr.String())
			racc.Contracts = map[string][]byte{
				tests.ContractHelloString.Name: tests.ContractHelloString.Source,
			}

			gw.GetAccount.Return(racc, nil)
		})

		di := &DependencyInstaller{
			Gateways: map[string]gateway.Gateway{
				config.EmulatorNetwork.Name: gw.Mock,
				config.TestnetNetwork.Name:  gw.Mock,
				config.MainnetNetwork.Name:  gw.Mock,
			},
			Logger: logger,
			State:  state,
		}

		err := di.Install()
		assert.NoError(t, err, "Failed to install dependencies")

		filePath := fmt.Sprintf("imports/%s/Hello", serviceAddress.String())
		fileContent, err := state.ReaderWriter().ReadFile(filePath)
		assert.NoError(t, err, "Failed to read generated file")
		assert.NotNil(t, fileContent)
	})
}
