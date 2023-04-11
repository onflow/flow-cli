/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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

package events

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

func Test_Get(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"test.event"}
		eventsFlags.Start = 10
		eventsFlags.End = 20

		result, err := get(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Success not passed start end", func(t *testing.T) {
		inArgs := []string{"test.event"}
		eventsFlags.Start = 0
		eventsFlags.End = 0

		srv.GetBlock.Run(func(args mock.Arguments) {
			query := args.Get(1).(flowkit.BlockQuery)
			assert.True(t, query.Latest)
		}).Return(tests.NewBlock(), nil)

		result, err := get(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail invalid range", func(t *testing.T) {
		inArgs := []string{"test.event"}
		eventsFlags.Start = 20
		eventsFlags.End = 0

		result, err := get(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "please provide either both start and end for range or only last flag")
		assert.Nil(t, result)
	})

}

func Test_Result(t *testing.T) {
	block := tests.NewBlock()
	event := EventResult{
		BlockEvents: []flow.BlockEvents{{
			BlockID:        block.ID,
			Height:         block.Height,
			BlockTimestamp: block.Timestamp,
			Events: []flow.Event{
				*tests.NewEvent(
					0,
					"A.foo",
					[]cadence.Field{{Type: cadence.StringType{}, Identifier: "bar"}},
					[]cadence.Value{cadence.NewInt(1)},
				),
			},
		}},
	}

	assert.Equal(t, strings.TrimPrefix(`
Events Block #1:
    Index	0
    Type	A.foo
    Tx ID	0000000000000000000000000000000000000000000000000000000000000000
    Values
		- bar (String): 1 

`, "\n"), event.String())

	assert.Equal(t, []any{map[string]any{
		"blockID":       uint64(1),
		"index":         0,
		"transactionId": "0000000000000000000000000000000000000000000000000000000000000000",
		"type":          "A.foo",
		"values":        json.RawMessage{0x7b, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x7b, 0x22, 0x69, 0x64, 0x22, 0x3a, 0x22, 0x41, 0x2e, 0x66, 0x6f, 0x6f, 0x22, 0x2c, 0x22, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x7b, 0x22, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x31, 0x22, 0x2c, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x22, 0x49, 0x6e, 0x74, 0x22, 0x7d, 0x2c, 0x22, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x3a, 0x22, 0x62, 0x61, 0x72, 0x22, 0x7d, 0x5d, 0x7d, 0x2c, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x22, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x7d, 0xa},
	}}, event.JSON())
}
