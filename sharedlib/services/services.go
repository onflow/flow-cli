package services

import (
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/gateway"
)

// Services are a place where we define domain functionality
type Services struct {
	Accounts     *Accounts
	Scripts      *Scripts
	Transactions *Transactions
	Keys         *Keys
	Events       *Events
}

// NewServices create new services with gateway and project
func NewServices(gateway gateway.Gateway, project cli.Project) *Services {
	return &Services{
		Accounts:     NewAccounts(gateway, project),
		Scripts:      NewScripts(gateway, project),
		Transactions: NewTransactions(gateway, project),
		Keys:         NewKeys(gateway, project),
		Events:       NewEvents(gateway, project),
	}
}
