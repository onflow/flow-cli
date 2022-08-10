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

package services

import (
	"fmt"
	"path"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/test-framework"
)

// Tests is a service that handles all tests-related interactions.
//
type Tests struct {
	logger output.Logger
}

// NewTests returns a new tests service.
//
func NewTests(
	logger output.Logger,
) *Tests {
	return &Tests{
		logger: logger,
	}
}

// Execute test scripts.
//
func (t *Tests) Execute(code []byte, scriptPath string, readerWriter flowkit.ReaderWriter) (test_framework.Results, error) {
	runner := test_framework.NewTestRunner().
		WithImportResolver(func(location common.Location) (string, error) {
			stringLocation, isFileImport := location.(common.StringLocation)
			if !isFileImport {
				return "", fmt.Errorf("cannot import from %s", location)
			}

			importFilePath := absolutePath(scriptPath, stringLocation.String())

			content, err := readerWriter.ReadFile(importFilePath)
			if err != nil {
				return "", err
			}

			return string(content), nil
		})

	t.logger.Info("Running tests...")

	return runner.RunTests(string(code))
}

func absolutePath(basePath, filePath string) string {
	if path.IsAbs(filePath) {
		return filePath
	}

	return path.Join(path.Dir(basePath), filePath)
}
