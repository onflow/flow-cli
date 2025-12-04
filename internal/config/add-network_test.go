/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

func setupTestState(t *testing.T) *flowkit.State {
	state, err := flowkit.Init(&afero.Afero{Fs: afero.NewMemMapFs()})
	require.NoError(t, err)
	return state
}

func TestAddNetworkWithFork(t *testing.T) {
	state := setupTestState(t)
	state.Networks().AddOrUpdate(config.Network{Name: "mainnet", Host: "access.mainnet.nodes.onflow.org:9000"})

	addNetworkFlags = flagsAddNetwork{Name: "forked-mainnet", Host: "http://127.0.0.1:3569", Fork: "mainnet"}
	result, err := addNetwork(nil, command.GlobalFlags{ConfigPaths: []string{"flow.json"}}, output.NewStdoutLogger(output.NoneLog), nil, state)

	require.NoError(t, err)
	require.NotNil(t, result)

	forkedNet, _ := state.Networks().ByName("forked-mainnet")
	assert.Equal(t, "mainnet", forkedNet.Fork)
}

func TestAddNetworkWithFork_InvalidSource(t *testing.T) {
	state := setupTestState(t)

	addNetworkFlags = flagsAddNetwork{Name: "forked-custom", Host: "http://127.0.0.1:3569", Fork: "custom-network"}
	result, err := addNetwork(nil, command.GlobalFlags{ConfigPaths: []string{"flow.json"}}, output.NewStdoutLogger(output.NoneLog), nil, state)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fork network \"custom-network\" not found")
	assert.Nil(t, result)
}

func TestAddNetworkWithoutFork(t *testing.T) {
	state := setupTestState(t)

	addNetworkFlags = flagsAddNetwork{Name: "testnet", Host: "access.devnet.nodes.onflow.org:9000", Fork: ""}
	result, err := addNetwork(nil, command.GlobalFlags{ConfigPaths: []string{"flow.json"}}, output.NewStdoutLogger(output.NoneLog), nil, state)

	require.NoError(t, err)
	require.NotNil(t, result)

	testnetNet, _ := state.Networks().ByName("testnet")
	assert.Equal(t, "", testnetNet.Fork)
}
