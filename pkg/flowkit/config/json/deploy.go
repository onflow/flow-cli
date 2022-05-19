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

package json

import (
	"encoding/json"
	"fmt"

	"github.com/onflow/cadence"

	jsoncdc "github.com/onflow/cadence/encoding/json"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

type jsonDeployments map[string]jsonDeployment

// transformToConfig transforms json structures to config structure.
func (j jsonDeployments) transformToConfig() (config.Deployments, error) {
	deployments := make(config.Deployments, 0)

	for networkName, deploys := range j {

		var deploy config.Deployment
		for accountName, contracts := range deploys {
			deploy = config.Deployment{
				Network: networkName,
				Account: accountName,
			}

			var contractDeploys []config.ContractDeployment
			for _, contract := range contracts {
				if contract.simple != "" {
					contractDeploys = append(
						contractDeploys,
						config.ContractDeployment{
							Name: contract.simple,
							Args: nil,
						},
					)
				} else {
					args := make([]cadence.Value, 0)
					for _, arg := range contract.advanced.Args {
						b, err := json.Marshal(arg)
						if err != nil {
							return nil, err
						}

						cadenceArg, err := jsoncdc.Decode(b)
						if err != nil {
							return nil, err
						}

						args = append(args, cadenceArg)
					}

					contractDeploys = append(
						contractDeploys,
						config.ContractDeployment{
							Name: contract.advanced.Name,
							Args: args,
						},
					)
				}
			}

			deploy.Contracts = contractDeploys
			deployments = append(deployments, deploy)
		}
	}

	return deployments, nil
}

// transformToJSON transforms config structure to json structures for saving.
func transformDeploymentsToJSON(configDeployments config.Deployments) jsonDeployments {
	jsonDeploys := jsonDeployments{}

	for _, d := range configDeployments {

		deployments := make([]deployment, 0)
		for _, c := range d.Contracts {
			if len(c.Args) == 0 {
				deployments = append(deployments, deployment{
					simple: c.Name,
				})
			} else {
				args := make([]map[string]interface{}, 0)
				for _, arg := range c.Args {
					switch arg.Type().ID() {
					case "Bool":
						args = append(args, map[string]interface{}{
							"type":  arg.Type().ID(),
							"value": arg.ToGoValue(),
						})
					default:
						args = append(args, map[string]interface{}{
							"type":  arg.Type().ID(),
							"value": fmt.Sprintf("%v", arg.ToGoValue()),
						})
					}
				}

				deployments = append(deployments, deployment{
					advanced: contractDeployment{
						Name: c.Name,
						Args: args,
					},
				})
			}
		}

		if _, ok := jsonDeploys[d.Network]; ok {
			jsonDeploys[d.Network][d.Account] = deployments
		} else {
			jsonDeploys[d.Network] = jsonDeployment{
				d.Account: deployments,
			}
		}

	}

	return jsonDeploys
}

type contractDeployment struct {
	Name string                   `json:"name"`
	Args []map[string]interface{} `json:"args"`
}

type deployment struct {
	simple   string
	advanced contractDeployment
}

type jsonDeployment map[string][]deployment

func (d *deployment) UnmarshalJSON(b []byte) error {

	// simple format
	var simple string
	err := json.Unmarshal(b, &simple)
	if err == nil {
		d.simple = simple
		return nil
	}

	// advanced format
	var advanced contractDeployment
	err = json.Unmarshal(b, &advanced)
	if err == nil {
		d.advanced = advanced
	} else {
		return err
	}

	return nil
}

func (d deployment) MarshalJSON() ([]byte, error) {
	if d.simple != "" {
		return json.Marshal(d.simple)
	} else {
		return json.Marshal(d.advanced)
	}
}
