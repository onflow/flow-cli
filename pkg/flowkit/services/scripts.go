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

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

// Scripts is a service that handles all script-related interactions.
type Scripts struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewScripts returns a new scripts service.
func NewScripts(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Scripts {
	return &Scripts{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}

// Execute script code with passed arguments on the selected network.
func (s *Scripts) Execute(script *flowkit.Script, network string, query *util.ScriptQuery) (cadence.Value, error) {
	program, err := project.NewProgram(script)
	if err != nil {
		return nil, err
	}

	if program.HasImports() {
		contracts, err := s.state.DeploymentContractsByNetwork(network)
		if err != nil {
			return nil, err
		}

		importReplacer := project.NewImportReplacer(
			contracts,
			s.state.AliasesForNetwork(network),
		)

		if s.state == nil {
			return nil, config.ErrDoesNotExist
		}
		if network == "" {
			return nil, fmt.Errorf("missing network, specify which network to use to resolve imports in script code")
		}
		if script.Location() == "" {
			return nil, fmt.Errorf("resolving imports in scripts not supported")
		}

		program, err = importReplacer.Replace(program)
		if err != nil {
			return nil, err
		}
	}

	return s.gateway.ExecuteScript(program.Code(), script.Args, query)
}
