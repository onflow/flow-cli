package services

import (
	"github.com/onflow/flow-cli/flow/gateway"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/util"
)

// Services are a place where we define domain functionality
type Services struct {
	Accounts     *Accounts
	Scripts      *Scripts
	Transactions *Transactions
	Keys         *Keys
	Events       *Events
	Collections  *Collections
	Project      *Project
	Blocks       *Blocks
}

// NewServices create new services with gateway and project
func NewServices(
	gateway gateway.Gateway,
	project *lib.Project,
	logger util.Logger,
) *Services {
	return &Services{
		Accounts:     NewAccounts(gateway, project, logger),
		Scripts:      NewScripts(gateway, project, logger),
		Transactions: NewTransactions(gateway, project, logger),
		Keys:         NewKeys(gateway, project, logger),
		Events:       NewEvents(gateway, project, logger),
		Collections:  NewCollections(gateway, project, logger),
		Project:      NewProject(gateway, project, logger),
		Blocks:       NewBlocks(gateway, project, logger),
	}
}
