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
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flowkit/output"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestCreate(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	projectName := "foobar"

	// Test project creation
	_, err := create([]string{projectName}, command.GlobalFlags{}, logger, nil, state)
	assert.NoError(t, err, "Failed to create project")

	// Check project was created in target directory which is in pwd
	pwd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")

	target := filepath.Join(pwd, fmt.Sprintf("%s/flow.json", projectName))

	_, err = state.ReaderWriter().ReadFile(target)
	assert.NoError(t, err, "Failed to read generated file")
}
