package services

import (
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/gateway"
	"github.com/onflow/flow-go-sdk"
)

// Collections service handles all interactions for collections
type Collections struct {
	gateway gateway.Gateway
	project *cli.Project
}

// NewCollections create new collection service
func NewCollections(gateway gateway.Gateway, project *cli.Project) *Collections {
	return &Collections{
		gateway: gateway,
		project: project,
	}
}

// Get collection
func (c *Collections) Get(id string) (*flow.Collection, error) {
	collectionID := flow.HexToID(id)
	return c.gateway.GetCollection(collectionID)
}
