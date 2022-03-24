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

package services

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/tests"
)

func TestScripts(t *testing.T) {
	t.Parallel()

	t.Run("Execute Script", func(t *testing.T) {
		_, s, gw := setup()

		gw.ExecuteScript.Run(func(args mock.Arguments) {
			assert.Equal(t, len(string(args.Get(0).([]byte))), 78)
			assert.Equal(t, args.Get(1).([]cadence.Value)[0].String(), "\"Foo\"")
			gw.ExecuteScript.Return(cadence.MustConvertValue(""), nil)
		})

		args := []cadence.Value{cadence.String("Foo")}
		_, err := s.Scripts.Execute(tests.ScriptArgString.Source, args, "", "")

		assert.NoError(t, err)
	})

}

func TestScripts_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Execute", func(t *testing.T) {
		t.Parallel()
		_, s := setupIntegration()

		args := []cadence.Value{cadence.String("Foo")}
		res, err := s.Scripts.Execute(tests.ScriptArgString.Source, args, "", "")

		assert.NoError(t, err)
		assert.Equal(t, res.String(), "\"Hello Foo\"")
	})

	t.Run("Execute report error", func(t *testing.T) {
		t.Parallel()
		_, s := setupIntegration()
		args := []cadence.Value{cadence.String("Foo")}
		res, err := s.Scripts.Execute(tests.ScriptWithError.Source, args, "", "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find type in this scope")
		assert.Nil(t, res)

	})

	t.Run("Execute With Imports", func(t *testing.T) {
		t.Parallel()
		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		// setup
		c := config.Contract{
			Name:    tests.ContractHelloString.Name,
			Source:  tests.ContractHelloString.Filename,
			Network: "emulator",
		}
		state.Contracts().AddOrUpdate(c.Name, c)

		n := config.Network{
			Name: "emulator",
			Host: "127.0.0.1:3569",
		}
		state.Networks().AddOrUpdate(n.Name, n)

		d := config.Deployment{
			Network: n.Name,
			Account: srvAcc.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)
		_, _ = s.Accounts.AddContract(srvAcc, tests.ContractHelloString.Name, tests.ContractHelloString.Source, false)

		res, err := s.Scripts.Execute(tests.ScriptImport.Source, nil, tests.ScriptImport.Filename, n.Name)
		assert.NoError(t, err)
		assert.Equal(t, res.String(), "\"Hello Hello, World!\"")
	})

	t.Run("Execute Script Invalid", func(t *testing.T) {
		t.Parallel()
		_, s := setupIntegration()
		in := [][]string{
			{tests.ScriptImport.Filename, ""},
			{"", "emulator"},
			{tests.ScriptImport.Filename, "foo"},
		}

		out := []string{
			"missing network, specify which network to use to resolve imports in script code",
			"resolving imports in scripts not supported",
			"import ./contractHello.cdc could not be resolved from the configuration",
		}

		for x, i := range in {
			_, err := s.Scripts.Execute(tests.ScriptImport.Source, nil, i[0], i[1])
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), out[x])
		}

	})
}
