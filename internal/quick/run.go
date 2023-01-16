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

package quick

import (
	"fmt"
	"sync"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/emulator"
	"github.com/onflow/flow-cli/internal/project"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"

	"github.com/spf13/cobra"
)

type flagsRun struct {
}

var runFlags = flagsRun{}

func DeployHelper(args []string, globalFlags command.GlobalFlags, services *services.Services, wg *sync.WaitGroup) {

	for {
		//check if the server has started
		_, err := services.Status.Ping(globalFlags.Network)
		if err == nil {
			// if the emulator is running run the deploy command
			project.DeployCommand.Cmd.Run(project.DeployCommand.Cmd, args)
			break
		}
	}

	wg.Done()

}

func EmulatorHelper(args []string, globalFlags command.GlobalFlags, services *services.Services, wg *sync.WaitGroup) {
	// run the emulator
	emulator.Cmd.Run(emulator.Cmd, args)
	wg.Done()
}

// RunCommand This command will act as an alias for running the emulator and deploying the contracts
var RunCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "run",
		Short:   "Start emulator and deploy all project contracts",
		Example: "flow run",
	},
	Flags: &runFlags,
	Run: func(
		args []string,
		_ flowkit.ReaderWriter,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		fmt.Println("⚠️Deprecation notice: Use 'flow dev' command.")
		return &RunResult{}, nil
	},
}

type RunResult struct{}

func (r *RunResult) JSON() interface{} {
	result := make(map[string]string)
	return result
}

func (r *RunResult) String() string {
	return ""
}

func (r *RunResult) Oneliner() string {
	return ""
}
