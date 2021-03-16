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
	"github.com/onflow/flow-cli/flow/gateway"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/util"
	"github.com/onflow/flow-go-sdk"
)

// Collections service handles all interactions for collections
type Collections struct {
	gateway gateway.Gateway
	project *lib.Project
	logger  util.Logger
}

// NewCollections create new collection service
func NewCollections(
	gateway gateway.Gateway,
	project *lib.Project,
	logger util.Logger,
) *Collections {
	return &Collections{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Get collection
func (c *Collections) Get(id string) (*flow.Collection, error) {
	collectionID := flow.HexToID(id)
	return c.gateway.GetCollection(collectionID)
}
