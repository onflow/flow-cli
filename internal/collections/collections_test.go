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
