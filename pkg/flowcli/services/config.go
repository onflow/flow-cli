package services

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
)

// Accounts is a service that handles all account-related interactions.
type Config struct {
	gateway gateway.Gateway
	project *project.Project
	logger  output.Logger
}

// NewAccounts returns a new accounts service.
func NewConfig(
	gateway gateway.Gateway,
	project *project.Project,
	logger output.Logger,
) *Config {
	return &Config{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

func (c *Config) AddAccount(account config.Account) error {
	if c.project == nil {
		return fmt.Errorf("missing configuration, initialize it: flow init")
	}

	c.project.Config().Accounts.AddOrUpdate(account.Name, account)

	return c.project.SaveDefault()
}

func (c *Config) RemoveAccount(name string) {

}

func (c *Config) AddContracts(contracts []config.Contract) error {
	if c.project == nil {
		return fmt.Errorf("missing configuration, initialize it: flow init")
	}

	for _, contract := range contracts {
		c.project.Config().Contracts.AddOrUpdate(contract.Name, contract)
	}

	return c.project.SaveDefault()
}

func (c *Config) RemoveContract(name string) {

}

func (c *Config) AddNetwork(network config.Network) error {
	if c.project == nil {
		return fmt.Errorf("missing configuration, initialize it: flow init")
	}

	c.project.Config().Networks.AddOrUpdate(network.Name, network)

	return c.project.SaveDefault()
}

func (c *Config) RemoveNetwork(name string) {

}

func (c *Config) AddDeployment(deployment config.Deploy) error {
	if c.project == nil {
		return fmt.Errorf("missing configuration, initialize it: flow init")
	}

	c.project.Config().Deployments.AddOrUpdate(deployment)

	return c.project.SaveDefault()
}

func (c *Config) RemoveDeployment(name string) {

}

func (c *Config) ListDeployments() {

}

func (c *Config) ListAccounts() {

}

func (c *Config) ListNetworks() {

}

func (c *Config) ListContracts() {

}
