package services

import (
	"fmt"
	"strconv"

	"github.com/onflow/flow-cli/sharedlib/util"

	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/gateway"
)

// Events service handles all interactions for scripts
type Events struct {
	gateway gateway.Gateway
	project *cli.Project
	logger  util.Logger
}

// NewEvents create new event service
func NewEvents(
	gateway gateway.Gateway,
	project *cli.Project,
	logger util.Logger,
) *Events {
	return &Events{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Get an event
func (e *Events) Get(name string, start string, end string) ([]client.BlockEvents, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("cannot use empty string as event name")
	}

	startHeight, err := strconv.ParseUint(start, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start height of block range: %v", startHeight)
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

	return e.gateway.GetEvents(name, startHeight, endHeight)
}
