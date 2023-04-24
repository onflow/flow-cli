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

package config

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"
)

func Test_Deployment(t *testing.T) {
	contracts := []ContractDeployment{{
		Name: "contract-1",
		Args: []cadence.Value{cadence.NewInt(5)},
	}, {
		Name: "contract-2",
		Args: []cadence.Value{cadence.NewInt(15)},
	}}

	t.Run("Adding deployments", func(t *testing.T) {
		network := "test-network"

		deployments := &Deployments{}
		deployments.AddOrUpdate(Deployment{
			Network:   network,
			Account:   "test-account",
			Contracts: []ContractDeployment{contracts[0]},
		})
		assert.Equal(t, []ContractDeployment{contracts[0]}, (*deployments)[0].Contracts)

		deployments.AddOrUpdate(Deployment{
			Network:   network,
			Account:   "test-account",
			Contracts: contracts,
		})
		assert.Equal(t, contracts, (*deployments)[0].Contracts)

		deployments.AddOrUpdate(Deployment{
			Network:   network,
			Account:   "test-account-2",
			Contracts: []ContractDeployment{contracts[0]},
		})
		assert.Len(t, *deployments, 2)
		assert.Equal(t, []ContractDeployment{contracts[0]}, (*deployments)[1].Contracts)
	})

	t.Run("Remove deployment", func(t *testing.T) {
		deployments := &Deployments{
			Deployment{
				Network: "test-network",
				Account: "test-account",
				Contracts: []ContractDeployment{{
					Name: "contract-1",
					Args: []cadence.Value{cadence.NewInt(5)},
				}, {
					Name: "contract-2",
					Args: []cadence.Value{cadence.NewInt(15)},
				}},
			},
		}

		err := deployments.Remove("test-account", "test-network")
		assert.NoError(t, err)
		assert.Len(t, *deployments, 0)
	})

	t.Run("Remove deployment contract", func(t *testing.T) {
		copyContracts := make([]ContractDeployment, len(contracts))
		copy(copyContracts, contracts)
		const acc = "test"
		const net = "testnet"

		deployments := &Deployments{
			Deployment{
				Network:   net,
				Account:   acc,
				Contracts: copyContracts,
			},
		}

		deployments.ByAccountAndNetwork(acc, net).RemoveContract(contracts[0].Name)

		assert.Len(t, *deployments, 1)
		assert.Len(t, (*deployments)[0].Contracts, 1)
		assert.Equal(t, (*deployments)[0].Contracts[0], contracts[1])
	})

	t.Run("Deployment by network", func(t *testing.T) {
		deployments := &Deployments{
			Deployment{
				Network: "net",
				Account: "acc",
			},
			Deployment{
				Network: "net",
				Account: "acc2",
			},
			Deployment{
				Network: "net2",
				Account: "acc2",
			},
		}

		network := deployments.ByNetwork("net")
		assert.Len(t, network, 2)
		assert.Equal(t, network[0].Account, "acc")
		assert.Equal(t, network[1].Account, "acc2")
	})

	t.Run("Remove non-existing deployment", func(t *testing.T) {
		deployments := &Deployments{
			Deployment{
				Network: "test",
				Account: "acc",
			},
		}

		err := deployments.Remove("acc", "no")
		assert.EqualError(t, err, "deployment for account acc on network no does not exist in configuration")
	})

	t.Run("Add deployment contract", func(t *testing.T) {
		const acc = "test"
		const net = "testnet"
		deployments := &Deployments{
			Deployment{Network: net, Account: acc},
		}

		deployments.ByAccountAndNetwork(acc, net).AddContract(contracts[0])
		deployments.ByAccountAndNetwork(acc, net).AddContract(contracts[1])

		assert.Len(t, *deployments, 1)
		assert.Len(t, (*deployments)[0].Contracts, 2)
		assert.Equal(t, (*deployments)[0].Contracts[1], contracts[1])
		assert.Equal(t, (*deployments)[0].Contracts[0], contracts[0])
	})

}
