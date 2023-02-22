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
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

func TestProject(t *testing.T) {
	t.Parallel()

	t.Run("Init Project", func(t *testing.T) {
		t.Parallel()

		st, s, _ := setup()
		pkey := tests.PrivKeys()[0]
		init, err := s.Project.Init(st.ReaderWriter(), false, false, crypto.ECDSA_P256, crypto.SHA3_256, pkey)
		assert.NoError(t, err)

		sacc, err := init.EmulatorServiceAccount()
		assert.NotNil(t, sacc)
		assert.NoError(t, err)
		assert.Equal(t, sacc.Name(), config.DefaultEmulatorServiceAccountName)
		assert.Equal(t, sacc.Address().String(), "f8d6e0586b0a20c7")

		p, err := sacc.Key().PrivateKey()
		assert.NoError(t, err)
		assert.Equal(t, (*p).String(), pkey.String())

		init, err = s.Project.Init(st.ReaderWriter(), false, false, crypto.ECDSA_P256, crypto.SHA3_256, nil)
		assert.NoError(t, err)
		em, err := init.EmulatorServiceAccount()
		assert.NoError(t, err)
		k, err := em.Key().PrivateKey()
		assert.NoError(t, err)
		assert.NotNil(t, (*k).String())
	})

	t.Run("Deploy Project", func(t *testing.T) {
		t.Parallel()

		state, s, gw := setup()

		c := config.Contract{
			Name:     "Hello",
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c.Name, c)

		n := config.Network{
			Name: "emulator",
			Host: "127.0.0.1:3569",
		}
		state.Networks().AddOrUpdate(n.Name, n)

		acct2 := tests.Donald()
		state.Accounts().AddOrUpdate(acct2)

		d := config.Deployment{
			Network: n.Name,
			Account: acct2.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			assert.Equal(t, tx.FlowTransaction().Payer, acct2.Address())
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		contracts, err := s.Project.Deploy("emulator", false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 1)
		assert.Equal(t, contracts[0].AccountAddress, acct2.Address())
	})

	t.Run("Deploy Project Using LocationAliases", func(t *testing.T) {
		t.Parallel()

		emulator := config.DefaultEmulatorNetwork().Name
		state, s, gw := setup()

		c1 := config.Contract{
			Name:     "ContractB",
			Location: tests.ContractB.Filename,
		}
		state.Contracts().AddOrUpdate(c1.Name, c1)

		c2 := config.Contract{
			Name:     "Aliased",
			Location: tests.ContractA.Filename,
			Aliases: []config.Alias{{
				Network: emulator,
				Address: tests.Donald().Address(),
			}},
		}
		state.Contracts().AddOrUpdate(c2.Name, c2)

		c3 := config.Contract{
			Name:     "ContractC",
			Location: tests.ContractC.Filename,
		}
		state.Contracts().AddOrUpdate(c3.Name, c3)

		state.Networks().AddOrUpdate(emulator, config.DefaultEmulatorNetwork())

		a := tests.Alice()
		state.Accounts().AddOrUpdate(a)

		d := config.Deployment{
			Network: emulator,
			Account: a.Name(),
			Contracts: []config.ContractDeployment{
				{Name: c1.Name}, {Name: c3.Name},
			},
		}
		state.Deployments().AddOrUpdate(d)

		// for checking imports are correctly resolved
		resolved := map[string]string{
			tests.ContractB.Name: fmt.Sprintf(`import ContractA from 0x%s`, tests.Donald().Address().Hex()),
			tests.ContractC.Name: fmt.Sprintf(`
		import ContractB from 0x%s
		import ContractA from 0x%s`, a.Address().Hex(), tests.Donald().Address().Hex()),
		} // don't change formatting of the above code since it compares the strings with included formatting

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			assert.Equal(t, tx.FlowTransaction().Payer, a.Address())
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			argCode := tx.FlowTransaction().Arguments[1]
			decodeCode, _ := jsoncdc.Decode(nil, argCode)
			code, _ := hex.DecodeString(decodeCode.ToGoValue().(string))

			argName := tx.FlowTransaction().Arguments[0]
			decodeName, _ := jsoncdc.Decode(nil, argName)

			testCode, found := resolved[decodeName.ToGoValue().(string)]
			require.True(t, found)
			assert.True(t, strings.Contains(string(code), testCode))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		contracts, err := s.Project.Deploy(emulator, false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 2)
		gw.Mock.AssertCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertCalled(t, tests.GetAccountFunc, a.Address())
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 2)
	})

	t.Run("Deploy Project New Import Schema and LocationAliases", func(t *testing.T) {
		t.Parallel()

		emulator := config.DefaultEmulatorNetwork().Name
		state, s, gw := setup()

		c1 := config.Contract{
			Name:     "ContractBB",
			Location: tests.ContractBB.Filename,
		}
		state.Contracts().AddOrUpdate(c1.Name, c1)

		c2 := config.Contract{
			Name:     "ContractAA",
			Location: tests.ContractAA.Filename,
			Aliases: []config.Alias{{
				Network: emulator,
				Address: tests.Donald().Address(),
			}},
		}
		state.Contracts().AddOrUpdate(c2.Name, c2)

		c3 := config.Contract{
			Name:     "ContractCC",
			Location: tests.ContractCC.Filename,
		}
		state.Contracts().AddOrUpdate(c3.Name, c3)

		state.Networks().AddOrUpdate(emulator, config.DefaultEmulatorNetwork())

		a := tests.Alice()
		state.Accounts().AddOrUpdate(a)

		d := config.Deployment{
			Network: emulator,
			Account: a.Name(),
			Contracts: []config.ContractDeployment{
				{Name: c1.Name}, {Name: c3.Name},
			},
		}
		state.Deployments().AddOrUpdate(d)

		// for checking imports are correctly resolved
		resolved := map[string]string{
			tests.ContractB.Name: fmt.Sprintf(`import ContractAA from 0x%s`, tests.Donald().Address().Hex()),
			tests.ContractC.Name: fmt.Sprintf(`
		import ContractBB from 0x%s
		import ContractAA from 0x%s`, a.Address().Hex(), tests.Donald().Address().Hex()),
		} // don't change formatting of the above code since it compares the strings with included formatting

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			assert.Equal(t, tx.FlowTransaction().Payer, a.Address())
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			argCode := tx.FlowTransaction().Arguments[1]
			decodeCode, _ := jsoncdc.Decode(nil, argCode)
			code, _ := hex.DecodeString(decodeCode.ToGoValue().(string))

			argName := tx.FlowTransaction().Arguments[0]
			decodeName, _ := jsoncdc.Decode(nil, argName)

			testCode, found := resolved[decodeName.ToGoValue().(string)]
			require.True(t, found)
			assert.True(t, strings.Contains(string(code), testCode))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		contracts, err := s.Project.Deploy(emulator, false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 2)
		gw.Mock.AssertCalled(t, tests.GetLatestBlockFunc)
		gw.Mock.AssertCalled(t, tests.GetAccountFunc, a.Address())
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 2)
	})

	t.Run("Deploy Project Duplicate Address", func(t *testing.T) {
		t.Parallel()

		state, s, gw := setup()

		c := config.Contract{
			Name:     "Hello",
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c.Name, c)

		n := config.Network{
			Name: "emulator",
			Host: "127.0.0.1:3569",
		}
		state.Networks().AddOrUpdate(n.Name, n)

		acct1 := tests.Charlie()
		state.Accounts().AddOrUpdate(acct1)

		acct2 := tests.Donald()
		state.Accounts().AddOrUpdate(acct2)

		d := config.Deployment{
			Network: n.Name,
			Account: acct2.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			assert.Equal(t, tx.FlowTransaction().Payer, acct2.Address())
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		contracts, err := s.Project.Deploy("emulator", false)

		assert.NoError(t, err)
		assert.Equal(t, len(contracts), 1)
		assert.Equal(t, contracts[0].AccountAddress, acct2.Address())
	})

}

// used for integration tests
func simpleDeploy(state *flowkit.State, s *Services, update bool) ([]*project.Contract, error) {
	srvAcc, _ := state.EmulatorServiceAccount()

	c := config.Contract{
		Name:     tests.ContractHelloString.Name,
		Location: tests.ContractHelloString.Filename,
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

	return s.Project.Deploy(n.Name, update)
}

func TestProject_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Deploy Project", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		contracts, err := simpleDeploy(state, s, false)
		assert.NoError(t, err)
		assert.Len(t, contracts, 1)
		assert.Equal(t, contracts[0].Name, tests.ContractHelloString.Name)
		assert.Equal(t, string(contracts[0].Code()), string(tests.ContractHelloString.Source))
	})

	t.Run("Deploy Complex Project", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		n := config.DefaultEmulatorNetwork()
		state.Networks().AddOrUpdate(n.Name, n)

		contractFixtures := []tests.Resource{
			tests.ContractB, tests.ContractC,
		}

		testContracts := make([]config.Contract, len(contractFixtures))
		for i, c := range contractFixtures {
			testContracts[i] = config.Contract{
				Name:     c.Name,
				Location: c.Filename,
			}
			state.Contracts().AddOrUpdate(c.Name, testContracts[i])
		}

		cA := tests.ContractA
		state.Contracts().AddOrUpdate(cA.Name, config.Contract{
			Name:     cA.Name,
			Location: cA.Filename,
			Aliases: []config.Alias{{
				Network: n.Name,
				Address: srvAcc.Address(),
			}},
		})

		state.Deployments().AddOrUpdate(config.Deployment{
			Network: n.Name,
			Account: srvAcc.Name(),
			Contracts: []config.ContractDeployment{{
				Name: testContracts[0].Name,
				Args: nil,
			}, {
				Name: testContracts[1].Name,
				Args: []cadence.Value{
					cadence.String("foo"),
				},
			}},
		})

		// deploy contract imported as alias
		_, _, err := s.Accounts.AddContract(
			srvAcc,
			flowkit.NewScript(tests.ContractA.Source, nil, tests.ContractA.Filename),
			n.Name,
			false,
		)
		require.NoError(t, err)

		// replace imports manually to assert that replacing worked in deploy service
		addr := fmt.Sprintf("0x%s", srvAcc.Address())
		replacedContracts := make([]string, len(contractFixtures))
		for i, c := range contractFixtures {
			replacedContracts[i] = strings.ReplaceAll(string(c.Source), `"./contractA.cdc"`, addr)
			replacedContracts[i] = strings.ReplaceAll(replacedContracts[i], `"./contractB.cdc"`, addr)
		}

		contracts, err := s.Project.Deploy(n.Name, false)
		assert.NoError(t, err)
		assert.Len(t, contracts, 2)

		account, err := s.Accounts.Get(srvAcc.Address())

		for i, c := range testContracts {
			code, exists := account.Contracts[c.Name]
			assert.True(t, exists)
			assert.Equal(t, replacedContracts[i], string(code))
		}
	})

	t.Run("Deploy Project Update", func(t *testing.T) {
		t.Parallel()

		// setup
		state, s := setupIntegration()
		_, err := simpleDeploy(state, s, false)
		assert.NoError(t, err)

		_, err = simpleDeploy(state, s, true)
		assert.NoError(t, err)
	})

}
