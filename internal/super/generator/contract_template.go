package generator

import (
	"fmt"
	"path/filepath"

	"github.com/onflow/flow-cli/internal/util"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
)

// Contract contains properties for contracts
type ContractTemplate struct {
	Name         string
	Account      string
	TemplatePath string
	Data         map[string]interface{}
	skipTests    bool
	saveState    bool
}

var _ TemplateItem = ContractTemplate{}

// GetTemplate returns the template of the contract
func (c ContractTemplate) GetTemplatePath() string {
	if c.TemplatePath == "" {
		return "contract_init.cdc.tmpl"
	}

	return c.TemplatePath
}

// GetData returns the data of the contract
func (c ContractTemplate) GetData() map[string]interface{} {
	data := map[string]interface{}{
		"Name": c.Name,
	}

	for k, v := range c.Data {
		data[k] = v
	}
	return data
}

func (c ContractTemplate) GetTargetPath() string {
	return filepath.Join(DefaultCadenceDirectory, "contracts", c.Account, util.AddCDCExtension(c.Name))
}

func (c ContractTemplate) UpdateState(state *flowkit.State) error {
	var aliases config.Aliases

	if c.skipTests != true {
		aliases = config.Aliases{{
			Network: config.TestingNetwork.Name,
			Address: flowsdk.HexToAddress("0x0000000000000007"),
		}}
	}

	contract := config.Contract{
		Name:     c.Name,
		Location: c.GetTargetPath(),
		Aliases:  aliases,
	}

	state.Contracts().AddOrUpdate(contract)

	if c.saveState {
		err := state.SaveDefault() // TODO: Support adding a target project directory
		if err != nil {
			return fmt.Errorf("error saving to flow.json: %w", err)
		}
	}

	return nil
}
