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

package vscode

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

const cadenceExt = "cadence.vsix"

var Cmd = &cobra.Command{
	Use:   "install-vscode-extension",
	Short: "Install the Cadence Visual Studio Code extension",
	Run: func(cmd *cobra.Command, args []string) {
		ext, _ := Asset(cadenceExt)

		// create temporary directory
		dir, err := ioutil.TempDir("", "vscode-cadence")
		if err != nil {
			util.Exit(1, err.Error())
		}

		// delete temporary directory
		defer os.RemoveAll(dir)

		tmpCadenceExt := fmt.Sprintf("%s/%s", dir, cadenceExt)

		err = ioutil.WriteFile(tmpCadenceExt, ext, 0644)
		if err != nil {
			util.Exit(1, err.Error())
		}

		// run vscode command to install extension from temporary directory
		c := exec.Command("code", "--install-extension", tmpCadenceExt)
		err = c.Run()
		if err != nil {
			util.Exit(1, err.Error())
		}

		fmt.Println("Installed the Cadence Visual Studio Code extension")
	},
}
