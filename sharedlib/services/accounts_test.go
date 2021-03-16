package services

import (
	"testing"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/sharedlib/util"
	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/sharedlib/tests"
)

func TestAccounts(t *testing.T) {
	serviceAddress := "f8d6e0586b0a20c7"
	mock := &tests.MockGateway{}
	project := cli.InitProject(crypto.ECDSA_P256, crypto.SHA3_256)

	accounts := NewAccounts(mock, project, util.NewStdoutLogger(util.InfoLog))

	t.Run("Get an Account", func(t *testing.T) {
		mock.GetAccountMock = func(address flow.Address) (*flow.Account, error) {
			return &flow.Account{Address: flow.HexToAddress(serviceAddress)}, nil
		}

		account, err := accounts.Get(serviceAddress)

		require.NoError(t, err)
		assert.Equal(t, account.Address.String(), serviceAddress)
	})

	// todo: write more mock tests
}
