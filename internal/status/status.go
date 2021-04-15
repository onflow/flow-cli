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

package status

import (
	"bytes"
	"fmt"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
	"github.com/spf13/cobra"
)

type FlagsStatus struct {
}

var statusFlags = FlagsStatus{}

var Command = &command.Command{
	Cmd: &cobra.Command{
		Use:   "status",
		Short: "Display the status of the Flow network",
	},
	Flags: &statusFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		// Get network name from global flag
		network := globalFlags.Network
		if network == "" {
			network = command.NetworkEmulator
		}

		accessNode, err := services.Status.Ping(network)

		return &Result{
			network:  network,
			accessNode: accessNode,
			err: err,
		}, nil
	},
}

type Result struct {
	network  string
	accessNode string
	err error
}


const (
	OnlineIcon   = "🟢"
	OnlineStatus = "ONLINE"

	OfflineIcon   = "🔴"
	OfflineStatus = "OFFLINE"
)

// getStatus returns string representation for Flow network status.
func (r *Result) getStatus() string {
	if r.err == nil {
		return OnlineStatus
	}

	return OfflineStatus
}

// getColoredStatus returns colored string representation for Flow network status.
func (r *Result) getColoredStatus() string {
	if r.err == nil {
		return util.Green(OnlineStatus)
	}

	return util.Red(OfflineStatus)
}

// getIcon returns emoji icon representing Flow network status.
func (r *Result) getIcon() string {
	if r.err == nil {
		return OnlineIcon
	}

	return OfflineIcon
}

// String converts result to a string.
func (r *Result) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	fmt.Fprintf(writer, "Status:\t %s %s\n", r.getIcon(), r.getColoredStatus())
	fmt.Fprintf(writer, "Network:\t %s\n", r.network)
	fmt.Fprintf(writer, "Access Node:\t %s\n", r.accessNode)

	writer.Flush()
	return b.String()
}

// JSON converts result to a JSON.
func (r *Result) JSON() interface{} {
	result := make(map[string]string)

	result["network"] = r.network
	result["accessNode"] = r.accessNode
	result["status"] = r.getStatus()

	return result
}

// Oneliner returns result as one liner grep friendly.
func (r *Result) Oneliner() string {
	return r.getStatus()
}
