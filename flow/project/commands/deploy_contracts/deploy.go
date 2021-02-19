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
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/flow/cli/txsender"
	"github.com/onflow/flow-cli/flow/project/contracts"
)

type Config struct {
	Network string `flag:"network" default:"emulator" info:"network configuration to use"`
	Update  bool   `flag:"update" default:"false" info:"use update flag to update existing contracts"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Cadence contracts",
	Run: func(cmd *cobra.Command, args []string) {
		project := cli.LoadProject()

		host := project.Host(conf.Network)

		grpcClient, err := client.New(host, grpc.WithInsecure())
		if err != nil {
			cli.Exit(1, err.Error())
			return
		}

		// check there are not multiple accounts with same contract
		if project.ContractConflictExists(conf.Network) {
			fmt.Println("\n❌ Currently it is not possible to deploy same contract with multiple accounts, please check Deploys in config and make sure contract is only present in one account")
			cli.Exit(1, "")
			return
		}

		sender := txsender.NewSender(grpcClient)

		processor := contracts.NewPreprocessor(
			contracts.FilesystemLoader{},
			project.GetAliases(conf.Network),
		)

		for _, contract := range project.GetContractsByNetwork(conf.Network) {
			err = processor.AddContractSource(
				contract.Name,
				contract.Source,
				contract.Target,
			)
			if err != nil {
				cli.Exit(1, err.Error())
				return
			}
		}

		err = processor.ResolveImports()
		if err != nil {
			cli.Exit(1, err.Error())
			return
		}

		contracts, err := processor.ContractDeploymentOrder()
		if err != nil {
			cli.Exit(1, err.Error())
			return
		}

		fmt.Printf(
			"Deploying %v contracts for accounts: %s\n",
			len(contracts),
			strings.Join(project.GetAllAccountNames(), ","),
		)

		var errs []error

		for _, contract := range contracts {
			targetAccount := project.GetAccountByAddress(contract.Target().String())

			ctx := context.Background()

			targetAccountInfo, err := grpcClient.GetAccountAtLatestBlock(ctx, targetAccount.Address())
			if err != nil {
				cli.Exitf(1, "Failed to fetch information for account %s", targetAccount.Address())
				return
			}

			var tx *flow.Transaction

			_, exists := targetAccountInfo.Contracts[contract.Name()]
			if exists {
				if !conf.Update {
					continue
				}

				tx = prepareUpdateContractTransaction(targetAccount.Address(), contract)
			} else {
				tx = prepareAddContractTransaction(targetAccount.Address(), contract)
			}

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

func prepareUpdateContractTransaction(
	targetAccount flow.Address,
	contract *contracts.Contract,
) *flow.Transaction {
	return templates.UpdateAccountContract(
		targetAccount,
		templates.Contract{
			Name:   contract.Name(),
			Source: contract.TranspiledCode(),
		},
	)
}

func prepareAddContractTransaction(
	targetAccount flow.Address,
	contract *contracts.Contract,
) *flow.Transaction {
	return templates.AddAccountContract(
		targetAccount,
		templates.Contract{
			Name:   contract.Name(),
			Source: contract.TranspiledCode(),
		},
	)
}
