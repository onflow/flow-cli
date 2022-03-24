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
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

// Services is a collection of services that provide domain-specific functionality
// for the different components of a Flow state.
type Services struct {
	Accounts     *Accounts
	Scripts      *Scripts
	Transactions *Transactions
	Keys         *Keys
	Events       *Events
	Collections  *Collections
	Project      *Project
	Blocks       *Blocks
	Status       *Status
	Snapshot     *Snapshot
}

// NewServices returns a new services collection for a state,
// initialized with a gateway and logger.
func NewServices(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Services {
	return &Services{
		Accounts:     NewAccounts(gateway, state, logger),
		Scripts:      NewScripts(gateway, state, logger),
		Transactions: NewTransactions(gateway, state, logger),
		Keys:         NewKeys(gateway, state, logger),
		Events:       NewEvents(gateway, state, logger),
		Collections:  NewCollections(gateway, state, logger),
		Project:      NewProject(gateway, state, logger),
		Blocks:       NewBlocks(gateway, state, logger),
		Status:       NewStatus(gateway, state, logger),
		Snapshot:     NewSnapshot(gateway, state, logger),
	}
}
