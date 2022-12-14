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
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_AccountFromPath(t *testing.T) {
	cadenceDir := "/Foo/cadence/"
	paths := [][]string{ // first is path, second is account name
		{"/Foo/cadence/contracts/alice/foo.cdc", "alice"},
		{"/Foo/cadence/contracts/alice", "alice"},
		{"/Foo/cadence/contracts/alice/boo/foo.cdc", ""},
		{"/Foo/cadence/contracts/foo.cdc", ""},
		{"/Foo/cadence/contracts/foo/bar/goo/foo", ""},
	}

	for i, test := range paths {
		name, ok := accountFromPath(cadenceDir, test[0])
		assert.Equal(t, test[1] != "", ok) // if we don't provide a name we mean it shouldn't be returned
		assert.Equal(t, test[1], name, fmt.Sprintf("failed test %d", i))
	}
}
