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

package super

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AccountFromPath(t *testing.T) {
	paths := [][]string{ // first is path, second is account name
		{"cadence/contracts/alice/foo.cdc", "alice"},
		{"cadence/contracts/alice", "alice"},
		{"cadence/contracts/alice/boo/foo.cdc", ""},
		{"cadence/contracts/foo.cdc", ""},
		{"cadence/contracts/foo/bar/goo/foo", ""},
	}

	for i, test := range paths {
		name, ok := accountFromPath(filepath.FromSlash(test[0]))
		assert.Equal(t, test[1] != "", ok) // if we don't provide a name we mean it shouldn't be returned
		assert.Equal(t, test[1], name, fmt.Sprintf("failed test %d", i))
	}
}

func Test_RelativeProjectPath(t *testing.T) {
	cdcDir := "/Users/Mike/Dev/my-project/cadence"
	paths := [][]string{
		{filepath.Join(cdcDir, "/contracts/foo.cdc"), "cadence/contracts/foo.cdc"},
		{filepath.Join(cdcDir, "/contracts/alice/foo.cdc"), "cadence/contracts/alice/foo.cdc"},
		{filepath.Join(cdcDir, "/scripts/bar.cdc"), "cadence/scripts/bar.cdc"},
		{filepath.Join(cdcDir, "/bar.cdc"), "cadence/bar.cdc"},
	}

	f := &projectFiles{
		cadencePath: cdcDir,
	}

	for i, test := range paths {
		rel, err := f.relProjectPath(filepath.FromSlash(test[0]))
		assert.NoError(t, err)
		assert.Equal(t, filepath.FromSlash(test[1]), rel, fmt.Sprintf("test %d failed", i))
	}
}
