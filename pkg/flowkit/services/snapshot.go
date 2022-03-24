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

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

// Snapshot is a service that handles downloading the latest finalized root protocol snapshot from the gateway.
type Snapshot struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewSnapshot returns a new snapshot service.
func NewSnapshot(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Snapshot {
	return &Snapshot{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}

// GetLatestProtocolStateSnapshot returns the latest finalized protocol snapshot
func (s *Snapshot) GetLatestProtocolStateSnapshot() ([]byte, error) {
	s.logger.StartProgress("Downloading protocol snapshot...")

	if !s.gateway.SecureConnection() {
		s.logger.Info(fmt.Sprintf("%s warning: using insecure client connection to download snapshot, you should use a secure network configuration...", output.WarningEmoji()))
	}

	b, err := s.gateway.GetLatestProtocolStateSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest finalized protocol snapshot from gateway: %w", err)
	}

	s.logger.StopProgress()

	return b, nil
}
