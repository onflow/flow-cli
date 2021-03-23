/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/config/output"
	"github.com/onflow/flow-cli/pkg/flow/util"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/pkg/flow/gateway"
)

// Scripts service handles all interactions for scripts
type Scripts struct {
	gateway gateway.Gateway
	project *flow.Project
	logger  output.Logger
}

// NewScripts create new script service
func NewScripts(
	gateway gateway.Gateway,
	project *flow.Project,
	logger output.Logger,
) *Scripts {
	return &Scripts{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Execute script
func (s *Scripts) Execute(scriptFilename string, args []string, argsJSON string) (cadence.Value, error) {
	script, err := util.LoadFile(scriptFilename)
	if err != nil {
		return nil, err
	}

	scriptArgs, err := flow.ParseArguments(args, argsJSON)
	if err != nil {
		return nil, err
	}

	return s.gateway.ExecuteScript(script, scriptArgs)
}
