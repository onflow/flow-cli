/*
 * Flow CLI
 *
 * Copyright 2024 Flow Foundation, Inc.
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
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/tests"
)

func Test_flix_generate(t *testing.T) {
	configJson := []byte(`{
		"contracts": {},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		},
		"networks": {
			"emulator": "127.0.0.1.3569"
		},
		"deployments": {
		}
	}`)

	af := afero.Afero{Fs: afero.NewMemMapFs()}
	err := afero.WriteFile(af.Fs, "flow.json", configJson, 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(af.Fs, tests.ContractHelloString.Filename, []byte(tests.ContractHelloString.Source), 0644)
	assert.NoError(t, err)
	paths := []string{"flow.json"}
	state, err := flowkit.Load(paths, af)
	assert.NotNil(t, state)
	assert.NoError(t, err)
	d := config.Deployment{
		Network: "emulator",
		Account: "emulator-account",
		Contracts: []config.ContractDeployment{{
			Name: tests.ContractHelloString.Name,
			Args: nil,
		}},
	}
	state.Deployments().AddOrUpdate(d)
	c := config.Contract{
		Name:     tests.ContractHelloString.Name,
		Location: tests.ContractHelloString.Filename,
	}
	state.Contracts().AddOrUpdate(c)

	contracts, err := state.DeploymentContractsByNetwork(config.Network{Name: "emulator"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(contracts))

	cs := GetDeployedContracts(state)
	assert.Equal(t, 1, len(cs))
	networkContract := cs[tests.ContractHelloString.Name]
	assert.NotNil(t, networkContract)
	addr := networkContract["emulator"]
	acct, err := state.Accounts().ByName("emulator-account")
	assert.NoError(t, err)
	assert.Equal(t, acct.Address.String(), addr)

}
