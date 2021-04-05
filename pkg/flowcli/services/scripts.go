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
	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/pkg/flowcli"
	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

// Scripts service handles all interactions for scripts
type Scripts struct {
	gateway gateway.Gateway
	project *project.Project
	logger  output.Logger
}

// NewScripts create new script service
func NewScripts(
	gateway gateway.Gateway,
	project *project.Project,
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

	return s.execute(script, args, argsJSON)
}

func (s *Scripts) ExecuteWithCode(code []byte, args []string, argsJSON string) (cadence.Value, error) {
	return s.execute(code, args, argsJSON)
}

func (s *Scripts) execute(code []byte, args []string, argsJSON string) (cadence.Value, error) {
	scriptArgs, err := flowcli.ParseArguments(args, argsJSON)
	if err != nil {
		return nil, err
	}

	return s.gateway.ExecuteScript(code, scriptArgs)
}
