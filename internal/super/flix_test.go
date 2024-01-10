package super

import (
	"testing"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/tests"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_flix_generate(t *testing.T) {
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

	cs := GetDeployedContracts(state)
	assert.Equal(t, 1, len(cs))

}
