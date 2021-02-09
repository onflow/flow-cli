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

package deploy_contracts

import (
	"context"
	"fmt"
	"log"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/onflow/flow-cli/flow/beta/cli"
	"github.com/onflow/flow-cli/flow/beta/contracts"
	"github.com/onflow/flow-cli/flow/project/cli/txsender"
)

type Config struct {
	Network string `flag:"network" default:"emulator" info:"network configuration to use"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "deploy-contracts",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		proj := cli.LoadProject()

		host := proj.Host(conf.Network)

		c, err := client.New(host, grpc.WithInsecure())
		if err != nil {
			cli.Exit(1, err.Error())
			return
		}

		sender := txsender.NewSender(c)

		p := contracts.NewPreprocessor(
			proj.Aliases(conf.Network),
			contracts.FilesystemResolver,
		)

		for _, contract := range proj.Contracts(conf.Network) {
			err = p.AddContractSource(
				contract.BundleName,
				contract.Name,
				contract.Source,
				contract.Target,
			)
			if err != nil {
				cli.Exit(1, err.Error())
				return
			}
		}

		contracts, err := p.PrepareForDeployment()
		if err != nil {
			cli.Exit(1, err.Error())
			return
		}

		fmt.Printf("Deploying contracts from: %s\n\n", cli.Yellow("nft, kitty-items"))

		var errs []error

		for _, contract := range contracts {

			targetAccount := proj.AccountByAddress(contract.Target())
			if targetAccount == nil {
				continue
			}

			tx := prepareDeploymentTransaction(targetAccount, contract)

			ctx := context.Background()

			getResult := sender.Send(ctx, tx, targetAccount)

			spinner := cli.NewSpinner(
				fmt.Sprintf("%s ", cli.Bold(contract.Name())),
				" deploying...",
			)
			spinner.Start()

			result := <-getResult

			if result.Error() == nil {
				spinner.Stop(fmt.Sprintf("%s -> 0x%s", cli.Green(contract.Name()), contract.Target()))
			} else {
				spinner.Stop(fmt.Sprintf("%s error", cli.Red(contract.Name())))
				errs = append(errs, result.Error())
			}
		}

		if len(errs) == 0 {
			fmt.Println("\n✅ All contracts deployed successfully")
		} else {
			fmt.Println("\n❌ Failed to deploy all contracts")
		}
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

func prepareDeploymentTransaction(
	targetAccount *cli.Account,
	contract *contracts.Contract,
) *flow.Transaction {

	return templates.AddAccountContracts(
		targetAccount.Address(),
		[]templates.Contract{
			{
				Name:   contract.Name(),
				Source: contract.TranspiledCode(),
			},
		},
		true,
	)
}
