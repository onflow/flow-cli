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
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"text/tabwriter"
)


const (
	OnlineIcon   = "🟢"
	OnlineStatus = "ONLINE"

	OfflineIcon   = "🔴"
	OfflineStatus = "OFFLINE"
)

// Status is a service that handles status of access node
type Status struct {
	gateway gateway.Gateway
	project *project.Project
	logger  output.Logger
}

// NewStatus returns a new ping service
func NewStatus(
	gateway gateway.Gateway,
	project *project.Project,
	logger output.Logger,
) *Status {
	return &Status{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Ping sends Ping request to network
func (s *Status) Ping(network string) PingResponse {
	err := s.gateway.Ping()
	accessNode := s.project.NetworkByName(network).Host

	return PingResponse{
		network: network,
		accessNode: accessNode,
		connectionError: err,
	}
}

type PingResponse struct {
	network string
	accessNode string
	connectionError error
}

// GetStatus returns string representation for network status
func (r *PingResponse) getStatus() string {
	if r.connectionError == nil {
		return color.GreenString("%s", OnlineStatus)
	}

	return color.RedString("%s", OfflineStatus)
}

// GetStatusIcon returns emoji icon representing network status
func (r *PingResponse) getStatusIcon() string {
	if r.connectionError == nil {
		return OnlineIcon
	}

	return OfflineIcon
}

func (r *PingResponse) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Status:\t %s %s\n", r.getStatusIcon(), r.getStatus())
	fmt.Fprintf(writer, "Network:\t %s\n", r.network)
	fmt.Fprintf(writer, "Access Node:\t %s\n", r.accessNode)

	writer.Flush()
	return b.String()
}

// JSON convert result to JSON
func (r *PingResponse) JSON() interface{} {
	result := make(map[string]string)

	result["network"] = r.network
	result["accessNode"] = r.accessNode
	result["status"] = r.getStatus()

	return result
}

// Oneliner show result as one liner grep friendly
func (r *PingResponse) Oneliner() string {
	return fmt.Sprintf("%s:%s", r.network, r.getStatus())
}

