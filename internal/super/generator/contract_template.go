/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

package generator

import (
	"fmt"
	"maps"
	"path/filepath"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"

	"github.com/onflow/flow-cli/internal/util"
)

const (
	DefaultContractDirectory = "contracts"
	DefaultTestAddress       = "0x0000000000000007"
)

// Contract contains properties for contracts
type ContractTemplate struct {
	Name         string
	FileName     string // Optional: If set, use this for the file name instead of Name
	Account      string
	TemplatePath string
	Data         map[string]any
	SkipTests    bool
	AddTestAlias bool
	Aliases      config.Aliases // Optional: Custom aliases for the contract
	SaveState    bool
}

var _ TemplateItem = ContractTemplate{}
var _ TemplateItemWithStateUpdate = ContractTemplate{}
var _ TemplateItemWithChildren = ContractTemplate{}

func (c ContractTemplate) GetType() string {
	return "contract"
}

func (c ContractTemplate) GetTemplatePath() string {
	if c.TemplatePath == "" {
		return "contract_init.cdc.tmpl"
	}

	return c.TemplatePath
}

func (c ContractTemplate) GetData() map[string]any {
	data := map[string]any{
		"Name": c.Name,
	}
	maps.Copy(data, c.Data)
	return data
}

func (c ContractTemplate) GetTargetPath() string {
	fileName := c.Name
	if c.FileName != "" {
		fileName = c.FileName
	}
	return filepath.Join(DefaultCadenceDirectory, DefaultContractDirectory, c.Account, util.AddCDCExtension(fileName))
}

func (c ContractTemplate) UpdateState(state *flowkit.State) error {
	var aliases config.Aliases

	// Use custom aliases if provided, otherwise use default test alias behavior
	if len(c.Aliases) > 0 {
		aliases = c.Aliases
	} else if c.SkipTests != true || c.AddTestAlias {
		aliases = config.Aliases{{
			Network: config.TestingNetwork.Name,
			Address: flowsdk.HexToAddress(DefaultTestAddress),
		}}
	}

	contract := config.Contract{
		Name:     c.Name,
		Location: c.GetTargetPath(),
		Aliases:  aliases,
	}

	state.Contracts().AddOrUpdate(contract)

	if c.SaveState {
		err := state.SaveDefault() // TODO: Support adding a target project directory
		if err != nil {
			return fmt.Errorf("error saving to flow.json: %w", err)
		}
	}

	return nil
}

func (c ContractTemplate) GetChildren() []TemplateItem {
	if c.SkipTests {
		return []TemplateItem{}
	}

	return []TemplateItem{
		TestTemplate{
			Name:         c.Name,
			TemplatePath: "contract_init_test.cdc.tmpl",
			Data: map[string]any{
				"ContractName": c.Name,
			},
		},
	}
}
