package collections

import (
	"github.com/onflow/flow-cli/internal/util"
	"testing"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Get(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{util.TestID.String()}

		srv.GetCollection.Run(func(args mock.Arguments) {
			id := args.Get(1).(flow.Identifier)
			assert.Equal(t, inArgs[0], id.String())
		})

		result, err := get(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}
