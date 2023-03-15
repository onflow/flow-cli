package project

import (
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ProjectDeploy(t *testing.T) {
	srv, state, rw := util.TestMocks(t)

	t.Run("Fail contract errors", func(t *testing.T) {
		srv.DeployProject.Return(nil, &flowkit.ProjectDeploymentError{})
		_, err := deploy([]string{}, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "failed deploying all contracts")
	})

	t.Run("Success replace standard contracts", func(t *testing.T) {
		const ft = "FungibleToken"
		state.Contracts().AddOrUpdate(config.Contract{
			Name:     ft,
			Location: "./ft.cdc",
		})
		_ = rw.WriteFile("./ft.cdc", []byte("test"), 0677) // mock the file

		state.Deployments().AddContract(
			config.DefaultEmulatorServiceAccountName,
			config.MainnetNetwork.Name,
			config.ContractDeployment{Name: "FungibleToken"},
		)

		err := checkForStandardContractUsageOnMainnet(state, util.NoLogger, true)
		require.NoError(t, err)

		assert.Len(t, state.Deployments().ByNetwork(config.EmulatorNetwork.Name), 0) // should remove it
		assert.NotNil(t, state.Contracts().ByName(ft).Aliases)
		assert.Equal(t, "f233dcee88fe0abe", state.Contracts().ByName(ft).Aliases.ByNetwork(config.MainnetNetwork.Name).Address.String())
	})

}
