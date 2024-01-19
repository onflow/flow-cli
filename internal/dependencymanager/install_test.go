package dependencymanager

import (
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/flowkit/tests"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestInstallDependencies(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	srv, state, _ := util.TestMocks(t)

	dep := config.Dependency{
		Name: "Hello",
		RemoteSource: config.RemoteSource{
			NetworkName:  "emulator",
			Address:      flow.HexToAddress("0000000000000001"),
			ContractName: "Hello",
		},
	}

	state.Dependencies().AddOrUpdate(dep)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"0x1"}

		srv.GetAccount.Run(func(args mock.Arguments) {
			addr := args.Get(1).(flow.Address)
			assert.Equal(t, "0000000000000001", addr.String())
			srv.GetAccount.Return(tests.NewAccountWithContracts(inArgs[0], tests.ContractHelloString), nil)
		})

		_, err := install([]string{}, command.GlobalFlags{}, logger, srv.Mock, state)
		assert.NoError(t, err, "Failed to install dependencies")

		fileContent, err := state.ReaderWriter().ReadFile("imports/0000000000000001/Hello")
		assert.NoError(t, err, "Failed to read generated file")
		assert.NotNil(t, fileContent)
	})
}
