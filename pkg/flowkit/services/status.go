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
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

// Status is a service that handles status of access node.
type Status struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewStatus returns a new status service.
func NewStatus(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Status {
	return &Status{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}

// Ping sends Ping request to network.
func (s *Status) Ping(network string) (string, error) {
	err := s.gateway.Ping()
	if err != nil {
		return "", err
	}
	n, err := s.state.Networks().ByName(network)
	if err != nil {
		return "", err
	}

	return n.Host, nil
}
