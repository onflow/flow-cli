package services

import (
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

	var events []client.BlockEvents
	if eventType != "" {
		events, err = e.gateway.GetEvents(eventType, block.Height, block.Height)
	}

	return block, events, err
}
