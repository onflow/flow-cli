/*
 * Flow CLI
 *
 * Copyright 2024 Flow Foundation, Inc.
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

package super

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flixkit-go/flixkit"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/mocks"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/tests"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type MockFlixService struct {
	mock.Mock
}

var TEMPLATE_STR = "{ \"f_type\": \"InteractionTemplate\", \"f_version\": \"1.1.0\", \"id\": \"0ea\",}"

func (m *MockFlixService) GetTemplate(ctx context.Context, templateName string) (string, string, error) {
	args := m.Called(ctx, templateName)
	return TEMPLATE_STR, args.String(0), args.Error(1)
}

var CADENCE_SCRIPT = "access(all) fun main() {\n    log(\"Hello, World!\")\n}"

func (m *MockFlixService) GetTemplateAndReplaceImports(ctx context.Context, templateName string, network string) (*flixkit.FlowInteractionTemplateExecution, error) {
	result := &flixkit.FlowInteractionTemplateExecution{
		Network:       "emulator",
		IsTransaciton: false,
		IsScript:      true,
		Cadence:       CADENCE_SCRIPT,
	}
	return result, nil
}

func (m *MockFlixService) CreateTemplate(ctx context.Context, contractInfos flixkit.ContractInfos, code string, preFill string, networks []config.Network) (string, error) {
	args := m.Called(ctx, contractInfos, code, preFill)
	return TEMPLATE_STR, args.Error(1)
}

var JS_CODE = "export async function request() { const info = await fcl.query({ template: flixTemplate }); return info; }"

func (m *MockFlixService) GetTemplateAndCreateBinding(ctx context.Context, templateName string, lang string, destFile string) (string, error) {
	args := m.Called(ctx, templateName, lang, destFile)
	return JS_CODE, args.Error(1)
}

func Test_ExecuteFlixScript(t *testing.T) {
	ctx := context.Background()
	logger := output.NewStdoutLogger(output.NoneLog)
	srv, state, _ := util.TestMocks(t)
	mockFlixService := new(MockFlixService)
	testCadenceScript := "access(all) fun main() {\n    log(\"Hello, World!\")\n}"
	mockFlixService.On("GetTemplateAndReplaceImports", ctx, "templateName", "emulator").Return(&flixkit.FlowInteractionTemplateExecution{
		Network:       "emulator",
		IsTransaciton: false,
		IsScript:      true,
		Cadence:       testCadenceScript,
	}, nil)

	// Set up a mock return value for the Network method
	mockNetwork := config.Network{
		Name: "emulator",
		Host: "localhost:3569",
	}
	srv.Network.Run(func(args mock.Arguments) {}).Return(mockNetwork, nil)
	srv.ExecuteScript.Run(func(args mock.Arguments) {
		script := args.Get(1).(flowkit.Script)
		assert.Equal(t, testCadenceScript, string(script.Code))
	}).Return(nil, nil)

	result, err := executeFlixCmd([]string{"transfer-token"}, command.GlobalFlags{}, logger, srv.Mock, state, mockFlixService)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func Test_ExecuteFlixTransaction(t *testing.T) {
	ctx := context.Background()
	logger := output.NewStdoutLogger(output.NoneLog)
	srv, state, _ := util.TestMocks(t)
	mockFlixService := new(MockFlixService)
	testCadenceTx := "transaction { prepare(signer: &Account) { /* prepare logic */ } execute { log(\"Hello, Cadence!\") } }"
	mockFlixService.On("GetTemplateAndReplaceImports", ctx, "templateName", "emulator").Return(&flixkit.FlowInteractionTemplateExecution{
		Network:       "emulator",
		IsTransaciton: false,
		IsScript:      true,
		Cadence:       testCadenceTx,
	}, nil)

	// Set up a mock return value for the Network method
	mockNetwork := config.Network{
		Name: "emulator",
		Host: "localhost:3569",
	}
	srv.Network.Run(func(args mock.Arguments) {}).Return(mockNetwork, nil)
	srv.SendTransaction.Run(func(args mock.Arguments) {
		script := args.Get(2).(flowkit.Script)
		assert.Equal(t, testCadenceTx, string(script.Code))
	}).Return(nil, nil)

	result, err := executeFlixCmd([]string{"transfer-token"}, command.GlobalFlags{}, logger, srv.Mock, state, mockFlixService)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func Test_PackageFlix(t *testing.T) {
	ctx := context.Background()
	logger := output.NewStdoutLogger(output.NoneLog)
	srv, state, _ := util.TestMocks(t)
	mockFlixService := new(MockFlixService)
	templateName := "templateName"
	mockFlixService.On("GetTemplateAndCreateBinding", ctx, templateName, "js", "").Return(JS_CODE, nil)

	result, err := packageFlixCmd([]string{templateName}, command.GlobalFlags{}, logger, srv.Mock, state, mockFlixService, flixFlags{Lang: "js"})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, JS_CODE, result.String())
}
func Test_GenerateFlix(t *testing.T) {
	srv := mocks.DefaultMockServices()
	cadenceFile := "cadence.cdc"
	cadenceCode := "access(all) fun main() {\n    log(\"Hello, World!\")\n}"

	mockFlixService := new(MockFlixService)

	configJson := []byte(`{
		"contracts": {},
		"accounts": {
			"emulator-account": {
				"address": "0xf8d6e0586b0a20c7",
				"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		},
		"networks": {
			"emulator": "127.0.0.1.3569"
		},
		"deployments": {
		}
	}`)

	af := afero.Afero{Fs: afero.NewMemMapFs()}
	err := afero.WriteFile(af.Fs, "flow.json", configJson, 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(af.Fs, cadenceFile, []byte(cadenceCode), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(af.Fs, tests.ContractHelloString.Filename, []byte(tests.ContractHelloString.Source), 0644)
	assert.NoError(t, err)
	paths := []string{"flow.json"}
	state, err := flowkit.Load(paths, af)
	assert.NotNil(t, state)
	assert.NoError(t, err)
	d := config.Deployment{
		Network: "emulator",
		Account: "emulator-account",
		Contracts: []config.ContractDeployment{{
			Name: tests.ContractHelloString.Name,
			Args: nil,
		}},
	}
	state.Deployments().AddOrUpdate(d)
	c := config.Contract{
		Name:     tests.ContractHelloString.Name,
		Location: tests.ContractHelloString.Filename,
	}
	state.Contracts().AddOrUpdate(c)

	contracts, err := state.DeploymentContractsByNetwork(config.Network{Name: "emulator"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(contracts))

	logger := output.NewStdoutLogger(output.NoneLog)
	contractInfos := make(flixkit.ContractInfos)
	contractInfos[tests.ContractHelloString.Name] = make(flixkit.NetworkAddressMap)
	contractInfos[tests.ContractHelloString.Name]["emulator"] = "0xf8d6e0586b0a20c7"

	ctx := context.Background()
	mockFlixService.On("CreateTemplate", ctx, contractInfos, cadenceCode, "").Return(TEMPLATE_STR, nil)

	result, err := generateFlixCmd([]string{cadenceFile}, command.GlobalFlags{}, logger, srv.Mock, state, mockFlixService, flixFlags{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TEMPLATE_STR, result.String())
}

func Test_GenerateFlixPrefill(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	srv := mocks.DefaultMockServices()
	templateName := "templateName"
	cadenceFile := "cadence.cdc"

	var mockFS = afero.NewMemMapFs()
	var rw = afero.Afero{Fs: mockFS}
	err := rw.WriteFile(cadenceFile, []byte(CADENCE_SCRIPT), 0644)
	assert.NoError(t, err)
	state, _ := flowkit.Init(rw)

	mockFlixService := new(MockFlixService)
	ctx := context.Background()
	mockFlixService.On("CreateTemplate", ctx, flixkit.ContractInfos{}, CADENCE_SCRIPT, templateName).Return(TEMPLATE_STR, nil)

	result, err := generateFlixCmd([]string{cadenceFile}, command.GlobalFlags{}, logger, srv.Mock, state, mockFlixService, flixFlags{PreFill: templateName})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
