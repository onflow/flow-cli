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
	"fmt"
	"strconv"

	"github.com/onflow/flow-go-sdk"
)

// GetBlock returns a block based on the provided query string.
//
// Query string options:
// - "latest"                : return the latest block
// - height (e.g. 123456789) : return block at this height
// - ID                      : return block with this ID
func (s *Services) GetBlock(
	query string,
	eventType string,
	verbose bool,
) (*flow.Block, []flow.BlockEvents, []*flow.Collection, error) {
	s.logger.StartProgress("Fetching Block...")
	defer s.logger.StopProgress()

	// smart parsing of query
	var err error
	var block *flow.Block
	if query == "latest" {
		block, err = s.gateway.GetLatestBlock()
	} else if height, ce := strconv.ParseUint(query, 10, 64); ce == nil {
		block, err = s.gateway.GetBlockByHeight(height)
	} else if flow.HexToID(query) != flow.EmptyID {
		block, err = s.gateway.GetBlockByID(flow.HexToID(query))
	} else {
		return nil, nil, nil, fmt.Errorf("invalid query: %s, valid are: \"latest\", block height or block ID", query)
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("error fetching block: %s", err.Error())
	}

	if block == nil {
		return nil, nil, nil, fmt.Errorf("block not found")
	}

	// if we specify event get events by the type
	var events []flow.BlockEvents
	if eventType != "" {
		events, err = s.gateway.GetEvents(eventType, block.Height, block.Height)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// if verbose fetch all collections from block too
	collections := make([]*flow.Collection, 0)
	if verbose {
		for _, guarantee := range block.CollectionGuarantees {
			collection, err := s.gateway.GetCollection(guarantee.CollectionID)
			if err != nil {
				return nil, nil, nil, err
			}
			collections = append(collections, collection)
		}
	}

	s.logger.StopProgress()

	return block, events, collections, err
}

// GetLatestBlockHeight returns the latest block height
func (s *Services) GetLatestBlockHeight() (uint64, error) {
	block, err := s.gateway.GetLatestBlock()
	if err != nil {
		return 0, err
	}
	return block.Height, nil
}
