package blocks

import (
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_GetBlock(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"100"}
		blockFlags.Events = "A.foo"
		blockFlags.Include = []string{"transactions"}

		srv.GetEvents.Run(func(args mock.Arguments) {
			assert.Equal(t, "A.foo", args.Get(1).([]string)[0])
			assert.Equal(t, 100, args.Get(2).(uint64))
			assert.Equal(t, 100, args.Get(3).(uint64))
		}).Return(nil, nil)

		srv.GetCollection.Return(nil, nil)

		srv.GetBlock.Run(func(args mock.Arguments) {
			assert.Equal(t, 100, args.Get(1).(flowkit.BlockQuery).Height)
		}).Return(tests.NewBlock(), nil)

		result, err := get(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.NotNil(t, result)
		assert.NoError(t, err)
	})
}
