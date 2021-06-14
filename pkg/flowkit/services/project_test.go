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

package services

import (
	"testing"

	"github.com/onflow/flow-cli/pkg/flowkit/config"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/tests"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
)

func TestProject(t *testing.T) {
	s, state, _, err := tests.ServicesStateMock()
	assert.NoError(t, err)
	project := s.Project

	t.Run("Init Project", func(t *testing.T) {
		s, err := flowkit.Init(state.ReaderWriter(), crypto.ECDSA_P256, crypto.SHA3_256)
		assert.NoError(t, err)

		sacc, err := s.EmulatorServiceAccount()
		assert.NotNil(t, sacc)
		assert.NoError(t, err)
	})

	t.Run("Deploy Project", func(t *testing.T) {
		c := config.Contract{
			Name:    "Hello",
			Source:  tests.ContractHelloString.Name,
			Network: "emulator",
		}
		state.Contracts().AddOrUpdate(c.Name, c)

		n := config.Network{
			Name: "emulator",
			Host: "127.0.0.1:3569",
		}
		state.Networks().AddOrUpdate(n.Name, n)

		a := tests.Alice()
		state.Accounts().AddOrUpdate(a)

		d := config.Deploy{
			Network: n.Name,
			Account: a.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)

		contracts, err := project.Deploy("emulator", false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 1)
	})
}
