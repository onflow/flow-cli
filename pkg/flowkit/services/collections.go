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
	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

// Collections is aa service that handles all collection-related interactions.
type Collections struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewCollections returns a new collections service.
func NewCollections(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Collections {
	return &Collections{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}

// Get returns a collection by ID.
func (c *Collections) Get(id flow.Identifier) (*flow.Collection, error) {
	return c.gateway.GetCollection(id)
}
