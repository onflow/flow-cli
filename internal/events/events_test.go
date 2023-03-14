package events

import (
	"github.com/onflow/flow-cli/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Get(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"test.event"}
		eventsFlags.Start = 10
		eventsFlags.End = 20

		result, err := get(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail invalid range", func(t *testing.T) {
		inArgs := []string{"test.event"}
		eventsFlags.Start = 20
		eventsFlags.End = 0

		result, err := get(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "please provide either both start and end for range or only last flag")
		assert.Nil(t, result)
	})

}
