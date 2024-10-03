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
	Account      string
	TemplatePath string
	Data         map[string]interface{}
	SkipTests    bool
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
	return filepath.Join(DefaultCadenceDirectory, DefaultContractDirectory, c.Account, util.AddCDCExtension(c.Name))
}

func (c ContractTemplate) UpdateState(state *flowkit.State) error {
	var aliases config.Aliases

	if c.SkipTests != true {
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
			Name:         fmt.Sprintf("%s_test", c.Name),
			TemplatePath: "contract_init_test.cdc.tmpl",
			Data: map[string]interface{}{
				"ContractName": c.Name,
			},
		},
	}
}
