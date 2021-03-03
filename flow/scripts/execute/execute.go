/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

package execute

import (
	"github.com/onflow/cadence"
	"io/ioutil"
	"log"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Flags struct {
	Args string `default:"" flag:"args" info:"arguments in JSON-Cadence format"`
	Code string `flag:"code,c" info:"path to Cadence file"`
	Host string `flag:"host" info:"Flow Access API host address"`
}

var flags Flags

var Cmd = &cobra.Command{
	Use:     "execute",
	Short:   "Execute a script",
	Example: `flow scripts execute --code=script.cdc --args="[{\"type\": \"String\", \"value\": \"Hello, Cadence\"}]"`,
	Run: func(cmd *cobra.Command, args []string) {
		code, err := ioutil.ReadFile(flags.Code)
		if err != nil {
			cli.Exitf(1, "Failed to read script from %s", flags.Code)
		}

		host, err := cli.LoadHostForNetwork(flags.Host, "")
		if err != nil {
			cli.Exitf(1, err.Error())
		}

		// Arguments
		var scriptArguments []cadence.Value
		if flags.Args != "" {
			scriptArguments, err = cli.ParseArguments(flags.Args)
			if err != nil {
				cli.Exitf(1, "Invalid arguments passed: %s", flags.Args)
			}
		}

		cli.ExecuteScript(host, code, scriptArguments...)
	},
}

func init() {
	initConfig()
}

func initConfig() {
	err := sconfig.New(&flags).
		FromEnvironment(cli.EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}
