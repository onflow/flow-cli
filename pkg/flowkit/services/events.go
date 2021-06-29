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
	"sync"

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

	startHeight, err := strconv.ParseUint(start, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start height of block range: %v", start)
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

	maxBlockRange := uint64(10000)
	if endHeight-startHeight > maxBlockRange {
		return nil, fmt.Errorf("block range is too big: %d, maximum block range is %d", endHeight-startHeight, maxBlockRange)
	}

	events, err := e.gateway.GetEvents(name, startHeight, endHeight)
	if err != nil {
		return nil, err
	}

	e.logger.StopProgress()
	return events, nil
}

func (e *Events) GetMany(events []string, start string, end string, blockCount uint64, workerCount int) ([]client.BlockEvents, error) {
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

	var queries []client.EventRangeQuery
	for startHeight <= endHeight {
		suggestedEndHeight := startHeight + blockCount
		endHeight := endHeight
		if suggestedEndHeight < endHeight {
			endHeight = suggestedEndHeight
		}
		for _, event := range events {
			queries = append(queries, client.EventRangeQuery{
				Type:        event,
				StartHeight: startHeight,
				EndHeight:   endHeight,
			})
		}
		startHeight = suggestedEndHeight + 1
	}

	jobChan := make(chan client.EventRangeQuery, workerCount)
	results := make(chan EventWorkerResult)

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eventWorker(jobChan, results, e)
		}()
	}

	// wait on the workers to finish and close the result channel
	// to signal downstream that all work is done
	go func() {
		defer close(results)
		wg.Wait()
	}()

	go func() {
		defer close(jobChan)
		for _, query := range queries {
			jobChan <- query
		}
	}()

	var resultEvents []client.BlockEvents
	for eventResult := range results {
		if eventResult.Error != nil {
			return nil, eventResult.Error
		}

		resultEvents = append(resultEvents, eventResult.Events...)
	}
	return resultEvents, nil

}

func eventWorker(jobChan <-chan client.EventRangeQuery, results chan<- EventWorkerResult, event *Events) {
	for q := range jobChan {
		blockEvents, err := event.Get(q.Type, strconv.FormatUint(q.StartHeight, 10), strconv.FormatUint(q.EndHeight, 10))
		if err != nil {
			results <- EventWorkerResult{nil, err}
		}
		results <- EventWorkerResult{blockEvents, nil}
	}
}


type EventWorkerResult struct {
	Events []client.BlockEvents
	Error  error
}