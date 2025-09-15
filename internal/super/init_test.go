/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ResolveTargetDirectory_CurrentDirCases(t *testing.T) {
	wd, err := filepath.Abs(".")
	assert.NoError(t, err)

	// Case 1: Empty input should resolve to current directory
	cur, err := resolveTargetDirectory("")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Clean(wd), filepath.Clean(cur))

	// Case 2: '.' should resolve to current directory
	dot, err := resolveTargetDirectory(".")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Clean(wd), filepath.Clean(dot))
}
