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
	"fmt"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type FlagsStatus struct {
}

var statusFlags = FlagsStatus{}

const (
	NetworkOnline = "ONLINE"
	NetworkOffline = "OFFLINE"
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
		err := services.Status.Ping()
		if err != nil {
			return &Result{NetworkOffline}, nil
		}

		return &Result{NetworkOnline}, nil
	},
}

// Result structure
type Result struct {
	status string
}

func (r *Result) String() string {
	icon := "🔴"
	if r.status == NetworkOnline {
		icon = "🟢"
	}
	return fmt.Sprintf("Network is: %s %s", icon, r.status)
}

// JSON convert result to JSON
func (r *Result) JSON() interface{} {
	return r
}

// Oneliner show result as one liner grep friendly
func (r *Result) Oneliner() string {
	return r.status
}
