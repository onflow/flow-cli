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

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
)

// Events service handles all interactions for scripts
type Events struct {
	gateway gateway.Gateway
	project *project.Project
	logger  output.Logger
}

// NewEvents create new event service
func NewEvents(
	gateway gateway.Gateway,
	project *project.Project,
	logger output.Logger,
) *Events {
	return &Events{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Get an event
func (e *Events) Get(name string, start string, end string) ([]client.BlockEvents, error) {
	if name == "" {
		return nil, fmt.Errorf("cannot use empty string as event name")
	}

	startHeight, err := strconv.ParseUint(start, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start height of block range: %v", start)
	}

	e.logger.StartProgress("Fetching Events...")

	var endHeight uint64
	if end == "" {
		endHeight = startHeight
	} else if end == "latest" {
		latestBlock, err := e.gateway.GetLatestBlock()
		if err != nil {
			return nil, err
		}

		endHeight = latestBlock.Height
	} else {
		endHeight, err = strconv.ParseUint(end, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end height of block range: %s", end)
		}
	}

	if endHeight < startHeight {
		return nil, fmt.Errorf("cannot have end height (%d) of block range less that start height (%d)", endHeight, startHeight)
	}

	maxBlockRange := uint64(10000)
	if endHeight-startHeight > maxBlockRange {
		return nil, fmt.Errorf("block range is too big: %d, maximum block range is %d", endHeight-startHeight, maxBlockRange)
	}

	events, err := e.gateway.GetEvents(name, startHeight, endHeight)
	if err != nil {
		return nil, err
	}

	e.logger.StopProgress("")
	return events, nil
}
