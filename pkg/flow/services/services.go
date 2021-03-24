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
	"github.com/onflow/flow-cli/pkg/flow/gateway"
	"github.com/onflow/flow-cli/pkg/flow/output"
	"github.com/onflow/flow-cli/pkg/flow/project"
)

// Services are a place where we define domain functionality
type Services struct {
	Accounts     *Accounts
	Scripts      *Scripts
	Transactions *Transactions
	Keys         *Keys
	Events       *Events
	Collections  *Collections
	Project      *Project
	Blocks       *Blocks
}

// NewServices create new services with gateway and project
func NewServices(
	gateway gateway.Gateway,
	project *project.Project,
	logger output.Logger,
) *Services {
	return &Services{
		Accounts:     NewAccounts(gateway, project, logger),
		Scripts:      NewScripts(gateway, project, logger),
		Transactions: NewTransactions(gateway, project, logger),
		Keys:         NewKeys(gateway, project, logger),
		Events:       NewEvents(gateway, project, logger),
		Collections:  NewCollections(gateway, project, logger),
		Project:      NewProject(gateway, project, logger),
		Blocks:       NewBlocks(gateway, project, logger),
	}
}
