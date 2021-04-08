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
	"github.com/fatih/color"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
	"text/tabwriter"
)

type FlagsStatus struct {
}

var statusFlags = FlagsStatus{}

// Result structure
type Result struct {
	network         string
	accessNode      string
	connectionError error
}

const (
	OnlineIcon   = "🟢"
	OnlineStatus = "ONLINE"

	OfflineIcon   = "🔴"
	OfflineStatus = "OFFLINE"
)

var InitCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "status",
		Short: "Display status of network",
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
			network = "emulator"
		}

		proj, err := project.Load(globalFlags.ConfigPath)

		if err != nil {
			return nil, fmt.Errorf("project can't be loaded from specified config path")
		}

		accessNode := proj.NetworkByName(network).Host

		err = services.Status.Ping()
		return &Result{
			network,
			accessNode,
			err,
		}, nil
	},
}

func (r *Result) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Status:\t %s %s\n", r.GetStatusIcon(), r.GetStatus())
	fmt.Fprintf(writer, "Network:\t %s\n", r.network)
	fmt.Fprintf(writer, "Access Node:\t %s\n", r.accessNode)

	writer.Flush()
	return b.String()
}

// JSON convert result to JSON
func (r *Result) JSON() interface{} {
	result := make(map[string]string)

	result["network"] = r.network
	result["accessNode"] = r.accessNode
	result["status"] = r.GetStatus()

	return result
}

// Oneliner show result as one liner grep friendly
func (r *Result) Oneliner() string {
	return fmt.Sprintf("%s:%s", r.network, r.GetStatus())
}

// GetStatus returns string representation for network status
func (r *Result) GetStatus() string {
	if r.connectionError == nil {
		return color.GreenString("%s", OnlineStatus)
	}

	return color.RedString("%s", OfflineStatus)
}

// GetStatusIcon returns emoji icon representing network status
func (r *Result) GetStatusIcon() string {
	if r.connectionError == nil {
		return OnlineIcon
	}

	return OfflineIcon
}
