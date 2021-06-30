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
	"strings"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

// Events is a service that handles all event-related interactions.
type Events struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewEvents returns a new events service.
func NewEvents(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Events {
	return &Events{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}

// Get queries for an event by name and block range.
func (e *Events) Get(name string, start string, end string) ([]client.BlockEvents, error) {
	if name == "" {
		return nil, fmt.Errorf("cannot use empty string as event name")
	}

	e.logger.StartProgress("Fetching Events...")
	defer e.logger.StopProgress()

	var err error
	var startHeight uint64
	if start == "latest" {
		latestBlock, err := e.gateway.GetLatestBlock()
		if err != nil {
			return nil, err
		}
		startHeight = latestBlock.Height

	} else if strings.HasPrefix(start, "-") {
		offset, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start height of block range: %v", start)
		}

		latestBlock, err := e.gateway.GetLatestBlock()
		if err != nil {
			return nil, err
		}
		startHeight = latestBlock.Height + uint64(offset)
	} else {
		startHeight, err = strconv.ParseUint(start, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start height of block range: %v", start)
		}
	}

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

	events, err := e.gateway.GetEvents(name, startHeight, endHeight)
	if err != nil {
		return nil, err
	}

	e.logger.StopProgress()
	return events, nil
}
