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

package status

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
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
	RunS:  status,
}

func status(
	_ []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
	_ *flowkit.State,
) (command.Result, error) {
	accessNode, err := services.Status.Ping(globalFlags.Network)

	return &Result{
		network:    globalFlags.Network,
		accessNode: accessNode,
		err:        err,
	}, nil
}

type Result struct {
	network    string
	accessNode string
	err        error
}

// getStatus returns string representation for Flow network status.
func (r *Result) getStatus() string {
	if r.err == nil {
		return "ONLINE"
	}

	return "OFFLINE"
}

// getColoredStatus returns colored string representation for Flow network status.
func (r *Result) getColoredStatus() string {
	if r.err == nil {
		return output.Green(r.getStatus())
	}

	return output.Red(output.Red(r.getStatus()))
}

// getIcon returns emoji icon representing Flow network status.
func (r *Result) getIcon() string {
	if r.err == nil {
		return output.GoEmoji()
	}

	return output.StopEmoji()
}

// String converts result to a string.
func (r *Result) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Status:\t %s %s\n", r.getIcon(), r.getColoredStatus())
	_, _ = fmt.Fprintf(writer, "Network:\t %s\n", r.network)
	_, _ = fmt.Fprintf(writer, "Access Node:\t %s\n", r.accessNode)

	_ = writer.Flush()
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
