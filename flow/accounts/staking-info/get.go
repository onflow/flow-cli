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

package staking_info

import (
	"fmt"
	"log"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-core-contracts/lib/go/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/onflow/flow-cli/flow"
)

type Config struct {
	Host string `flag:"host" info:"Flow Access API host address"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "staking-info <address>",
	Short: "Get account staking info",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectConf := new(cli.Config)
		if conf.Host == "" {
			projectConf = cli.LoadConfig()
		}

		address := flow.HexToAddress(args[0])

		cadenceAddress := cadence.NewAddress(address)

		chain, err := cli.GetAddressNetwork(address)

		if err != nil {
			cli.Exitf(1, "Failed to determine network from input address ")
		}

		env := cli.EnvFromNetwork(chain)

		stakingInfoScript := templates.GenerateGetLockedStakerInfoScript(env)

		fmt.Println("Account Staking Info:")
		cli.ExecuteScript(projectConf.HostWithOverride(conf.Host), []byte(stakingInfoScript), cadenceAddress)

		delegationInfoScript := templates.GenerateGetLockedDelegatorInfoScript(env)

		fmt.Println("Account Delegation Info:")
		cli.ExecuteScript(projectConf.HostWithOverride(conf.Host), []byte(delegationInfoScript), cadenceAddress)
	},
}

func init() {
	initConfig()
}

func initConfig() {
	err := sconfig.New(&conf).
		FromEnvironment(cli.EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}
