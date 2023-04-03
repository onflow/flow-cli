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

package blocks

import (
	"strings"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

func Test_GetBlock(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"100"}
		blockFlags.Events = "A.foo"
		blockFlags.Include = []string{"transactions"}

		srv.GetEvents.Run(func(args mock.Arguments) {
			assert.Equal(t, "A.foo", args.Get(1).([]string)[0])
			assert.Equal(t, uint64(100), args.Get(2).(uint64))
			assert.Equal(t, uint64(100), args.Get(3).(uint64))
		}).Return(nil, nil)

		srv.GetCollection.Return(nil, nil)

		returnBlock := tests.NewBlock()
		returnBlock.Height = uint64(100)

		srv.GetBlock.Run(func(args mock.Arguments) {
			assert.Equal(t, uint64(100), args.Get(1).(flowkit.BlockQuery).Height)
		}).Return(returnBlock, nil)

		result, err := get(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.NotNil(t, result)
		assert.NoError(t, err)
	})
}

func Test_Result(t *testing.T) {
	result := BlockResult{
		block:       tests.NewBlock(),
		collections: []*flow.Collection{tests.NewCollection()},
	}

	assert.Equal(t, strings.TrimPrefix(`
Block ID		0202020202020202020202020202020202020202020202020202020202020202
Parent ID		0303030303030303030303030303030303030303030303030303030303030303
Proposal Timestamp	2020-06-04 16:43:21 +0000 UTC
Proposal Timestamp Unix	1591289001
Height			1
Total Seals		1
Total Collections	3
    Collection 0:	0202020202020202020202020202020202020202020202020202020202020202
    Collection 1:	0303030303030303030303030303030303030303030303030303030303030303
    Collection 2:	0404040404040404040404040404040404040404040404040404040404040404
`, "\n"), result.String())

	assert.Equal(
		t,
		map[string]interface{}{
			"blockId": "0202020202020202020202020202020202020202020202020202020202020202",
			"collection": []interface{}{
				map[string]interface{}{"id": "0202020202020202020202020202020202020202020202020202020202020202"},
				map[string]interface{}{"id": "0303030303030303030303030303030303030303030303030303030303030303"},
				map[string]interface{}{"id": "0404040404040404040404040404040404040404040404040404040404040404"},
			},
			"height":           uint64(1),
			"parentId":         "0303030303030303030303030303030303030303030303030303030303030303",
			"totalCollections": 3,
			"totalSeals":       1,
		},
		result.JSON(),
	)
}
