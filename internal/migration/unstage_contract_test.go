package migration

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_UnstageContract(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	testContract := tests.ContractSimple

	t.Run("Success", func(t *testing.T) {

		srv.SendTransaction.Run(func(args mock.Arguments) {
			script := args.Get(2).(flowkit.Script)

			actualContractNameArg := script.Args[0]

			contractName, _ := cadence.NewString(testContract.Name)
			assert.Equal(t, contractName, actualContractNameArg)
		}).Return(flow.NewTransaction(), &flow.TransactionResult{
			Status:      flow.TransactionStatusSealed,
			Error:       nil,
			BlockHeight: 1,
		}, nil)

		result, err := unstageContract(
			[]string{testContract.Name},
			command.GlobalFlags{
				Network: "testnet",
			},
			util.NoLogger,
			srv.Mock,
			state,
		)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("missing contract file", func(t *testing.T) {
		result, err := stageContract(
			[]string{testContract.Name, "bad file path"},
			command.GlobalFlags{
				Network: "testnet",
			},
			util.NoLogger,
			srv.Mock,
			state,
		)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
