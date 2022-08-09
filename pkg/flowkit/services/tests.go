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
	"github.com/onflow/flow-cli/pkg/flowkit/output"

	"github.com/onflow/cadence/test-framework"
)

// Tests is a service that handles all tests-related interactions.
type Tests struct {
	logger output.Logger
}

// NewTests returns a new tests service.
func NewTests(
	logger output.Logger,
) *Tests {
	return &Tests{
		logger: logger,
	}
}

// Execute test scripts.
func (s *Tests) Execute(code []byte) (test_framework.Results, error) {
	runner := test_framework.NewTestRunner()
	return runner.RunTests(string(code))
}
