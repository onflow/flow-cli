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
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

// Services is a collection of services that provide domain-specific functionality
// for the different components of a Flow state.
// type Services struct {
// 	Accounts     *Accounts
// 	Scripts      *Scripts
// 	Transactions *Transactions
// 	Keys         *Keys
// 	Events       *Events
// 	Collections  *Collections
// 	Project      *Project
// 	Blocks       *Blocks
// 	Status       *Status
// 	Snapshot     *Snapshot
// }

type Services struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

type Flowkit interface {
	GetAccount(address flow.Address) (*flow.Account, error)
	StakingInfo(address flow.Address) ([]map[string]interface{}, []map[string]interface{}, error)
	NodeTotalStake(nodeId string, chain flow.ChainID) (*cadence.Value, error)
	CreateAddress(
		signer *flowkit.Account,
		pubKeys []crypto.PublicKey,
		keyWeights []int,
		sigAlgo []crypto.SignatureAlgorithm,
		hashAlgo []crypto.HashAlgorithm,
		contractArgs []string,
	) (*flow.Account, error)
	AddContract(
		account *flowkit.Account,
		contractName string,
		contractSource []byte,
		updateExisting bool,
		contractArgs []cadence.Value,
	) (*flow.Account, error)
	RemoveContract(
		account *flowkit.Account,
		contractName string,
	) (*flow.Account, error)
	GetBlock(
		query string,
		eventType string,
		verbose bool,
	) (*flow.Block, []flow.BlockEvents, []*flow.Collection, error)
	GetLatestBlockHeight() (uint64, error)
}

// NewServices returns a new services collection for a state,
// initialized with a gateway and logger.
func NewServices(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Services {
	return &Services{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}
