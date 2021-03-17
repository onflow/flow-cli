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
	"fmt"
	"strconv"

	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/flow/gateway"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/util"
	"github.com/onflow/flow-go-sdk"
)

// Blocks service handles all interactions for blocks
type Blocks struct {
	gateway gateway.Gateway
	project *lib.Project
	logger  util.Logger
}

// NewBlocks create new block service
func NewBlocks(
	gateway gateway.Gateway,
	project *lib.Project,
	logger util.Logger,
) *Blocks {
	return &Blocks{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Get the block
func (e *Blocks) GetBlock(query string, eventType string) (*flow.Block, []client.BlockEvents, error) {
	var block *flow.Block
	var err error

	if query == "latest" {
		block, err = e.gateway.GetLatestBlock()
	} else if height, err := strconv.ParseUint(query, 10, 64); err == nil {
		block, err = e.gateway.GetBlockByHeight(height)
	} else {
		block, err = e.gateway.GetBlockByID(flow.HexToID(query))
	}

	if block == nil {
		return nil, nil, fmt.Errorf("block not found")
	}

	var events []client.BlockEvents
	if eventType != "" {
		events, err = e.gateway.GetEvents(eventType, block.Height, block.Height)
	}

	return block, events, err
}
