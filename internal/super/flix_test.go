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
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/mocks"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/flowkit/tests"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type MockFlixService struct {
	mock.Mock
}

func (m *MockFlixService) GetTemplate(ctx context.Context, templateName string) (string, error) {
	args := m.Called(ctx, templateName)
	return args.String(0), args.Error(1)
}

func (m *MockFlixService) GetAndReplaceCadenceImports(ctx context.Context, templateName string, network string) (*flixkit.FlowInteractionTemplateExecution, error) {
	result := &flixkit.FlowInteractionTemplateExecution{
		Network:       "emulator",
		IsTransaciton: false,
		IsScript:      true,
		Cadence:       "pub fun main() {\n    log(\"Hello, World!\")\n}",
	}
	return result, nil
}
func Test_ExecuteFlixScript(t *testing.T) {
	ctx := context.Background()
	logger := output.NewStdoutLogger(output.NoneLog)
	srv, state, _ := util.TestMocks(t)
	mockFlixService := new(MockFlixService)
	testCadence := "pub fun main() {\n    log(\"Hello, World!\")\n}"
	mockFlixService.On("GetAndReplaceCadenceImports", ctx, "templateName", "emulator").Return(&flixkit.FlowInteractionTemplateExecution{
		Network:       "emulator",
		IsTransaciton: false,
		IsScript:      true,
		Cadence:       testCadence,
	}, nil)

	// Set up a mock return value for the Network method
	mockNetwork := config.Network{
		Name: "emulator",
		Host: "localhost:3569",
	}
	srv.Network.Run(func(args mock.Arguments) {}).Return(mockNetwork, nil)
	srv.ExecuteScript.Run(func(args mock.Arguments) {
		script := args.Get(1).(flowkit.Script)
		assert.Equal(t, testCadence, string(script.Code))
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
	testCadence := "transaction { prepare(signer: AuthAccount) { /* prepare logic */ } execute { log(\"Hello, Cadence!\") } }"
	mockFlixService.On("GetAndReplaceCadenceImports", ctx, "templateName", "emulator").Return(&flixkit.FlowInteractionTemplateExecution{
		Network:       "emulator",
		IsTransaciton: false,
		IsScript:      true,
		Cadence:       testCadence,
	}, nil)

	// Set up a mock return value for the Network method
	mockNetwork := config.Network{
		Name: "emulator",
		Host: "localhost:3569",
	}
	srv.Network.Run(func(args mock.Arguments) {}).Return(mockNetwork, nil)
	srv.SendTransaction.Run(func(args mock.Arguments) {
		script := args.Get(2).(flowkit.Script)
		assert.Equal(t, testCadence, string(script.Code))
	}).Return(nil, nil)

	result, err := executeFlixCmd([]string{"transfer-token"}, command.GlobalFlags{}, logger, srv.Mock, state, mockFlixService)
	assert.NoError(t, err)
	assert.NotNil(t, result)

}

type MockFclGenerator struct {
	mock.Mock
}

func (m *MockFclGenerator) Generate(flixString string, templateLocation string) (string, error) {
	args := m.Called(flixString, templateLocation)
	return args.String(0), args.Error(1)
}

func Test_PackageFlix(t *testing.T) {
	ctx := context.Background()
	logger := output.NewStdoutLogger(output.NoneLog)
	srv, state, _ := util.TestMocks(t)
	mockFlixService := new(MockFlixService)
	template := "{ \"f_type\": \"InteractionTemplate\", \"f_version\": \"1.1.0\", \"id\": \"0ea\",}"
	templateName := "templateName"
	mockFlixService.On("GetTemplate", ctx, templateName).Return(template, nil)
	jsCode := "export async function request() { const info = await fcl.query({ template: flixTemplate }); return info; }"
	mockFclGenerator := new(MockFclGenerator)
	mockFclGenerator.On("Generate", template, templateName).Return(jsCode, nil)

	result, err := packageFlixCmd([]string{templateName}, command.GlobalFlags{}, logger, srv.Mock, state, mockFlixService, mockFclGenerator)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, jsCode, result.String())
}

type MockGenerateTemplate struct {
	mock.Mock
}

func (m *MockGenerateTemplate) Generate(ctx context.Context, code string, preFill string) (string, error) {
	args := m.Called(ctx, code, preFill)
	return args.String(0), args.Error(1)
}
func Test_GenerateFlix(t *testing.T) {
	ctx := context.Background()
	srv := mocks.DefaultMockServices()
	template := "{ \"f_type\": \"InteractionTemplate\", \"f_version\": \"1.1.0\", \"id\": \"0ea\",}"
	templateName := "templateName"
	cadenceFile := "cadence.cdc"
	cadenceCode := "pub fun main() {\n    log(\"Hello, World!\")\n}"

	mockFlixService := new(MockFlixService)
	mockFlixService.On("GetTemplate", ctx, templateName).Return(template, nil)

	mockGenerateTemplate := new(MockGenerateTemplate)
	mockGenerateTemplate.On("Generate", cadenceCode, template).Return(template, nil)

	configJson := []byte(`{
		"contracts": {},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
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

	result, err := generateFlixCmd([]string{cadenceFile}, command.GlobalFlags{}, logger, srv.Mock, state, mockFlixService, mockGenerateTemplate, flixFlags{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, template, result.String())
}

func Test_GenerateFlixPrefill(t *testing.T) {
	ctx := context.Background()
	logger := output.NewStdoutLogger(output.NoneLog)
	srv := mocks.DefaultMockServices()
	template := "{ \"f_type\": \"InteractionTemplate\", \"f_version\": \"1.1.0\", \"id\": \"0ea\",}"
	templateName := "templateName"
	cadenceFile := "cadence.cdc"
	cadenceCode := "pub fun main() {\n    log(\"Hello, World!\")\n}"

	var mockFS = afero.NewMemMapFs()
	var rw = afero.Afero{Fs: mockFS}
	err := rw.WriteFile(cadenceFile, []byte(cadenceCode), 0644)
	assert.NoError(t, err)
	state, _ := flowkit.Init(rw, crypto.ECDSA_P256, crypto.SHA3_256)

	mockFlixService := new(MockFlixService)
	mockFlixService.On("GetTemplate", ctx, templateName).Return(template, nil)

	mockGenerateTemplate := new(MockGenerateTemplate)
	mockGenerateTemplate.On("Generate", ctx, cadenceCode, template).Return(template, nil)

	result, err := generateFlixCmd([]string{cadenceFile}, command.GlobalFlags{}, logger, srv.Mock, state, mockFlixService, mockGenerateTemplate, flixFlags{PreFill: templateName})
	assert.NoError(t, err)
	assert.NotNil(t, result)

}
