package services

import (
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/gateway"
)

type Services struct {
	Accounts *Accounts
}

func NewServices(gateway gateway.Gateway, project cli.Project) *Services {
	return &Services{
		Accounts: NewAccounts(gateway, project),
	}
}
