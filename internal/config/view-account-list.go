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

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsViewAccount struct {
}

var viewAccountFlags = flagsViewAccount{}

var ViewAccountCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "account",
		Short:   "View a list of accounts in configuration",
		Example: "flow config view account",
		Args:    cobra.NoArgs,
	},
	Flags: &viewAccountFlags,
	RunS:  viewAccount,
}

func viewAccount(args []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	{

		// Open our Configuration (flow.json File) File
		jsonFile, err := os.Open(config.DefaultPath)
		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		// Unmarshalling our json data into result
		var result map[string]interface{}
		json.Unmarshal([]byte(byteValue), &result)

		//  creating a temp variablle to handle json data in nested loop.
		//  Eg: For result map[string]interface{} => string -- emulators, contracts, accounts, deployments ;
		//  	 interface{} -- {
		// 					"emulator-account": {
		// 						"address": "f8d6e0586b0a20c7",
		// 						"key": "e8faab14240e6bb44b3713a21b41794f4f91301634acb0ec05339c3be0a3abb0"
		// 					}
		// 				}
		// 		For temp map[string]interface{} => string -- emulator-account(basically all account names);
		//   	  interface{} -- {
		// 					"address": "f8d6e0586b0a20c7",
		// 					"key": "e8faab14240e6bb44b3713a21b41794f4f91301634acb0ec05339c3be0a3abb0"
		// 				}

		var temp map[string]interface{}

		for key, value := range result {
			// checking if the key is accounts
			if key == "accounts" {
				// storing the value of the key -- accounts into temp for future iterations
				temp = value.(map[string]interface{})
			}
		}

		// Printing all account names
		fmt.Print("List of Accounts: ")
		for key, _ := range temp {
			fmt.Print(key, ", ")
		}
	}

	return &Result{}, nil
}
