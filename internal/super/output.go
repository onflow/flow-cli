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

package super

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"math/rand"
	"os"
	sysExec "os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	flowkitProject "github.com/onflow/flow-cli/pkg/flowkit/project"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

func printDeployment(deployed []*flowkitProject.Contract, err error, contractPathNames map[string]string) {
	clearScreen()
	fmt.Println(helpBanner())

	if err != nil {
		fmt.Println(errorBanner())
		fmt.Println(failureDeployment(err, contractPathNames))
		return
	}

	fmt.Println(okBanner())
	fmt.Println(successfulDeployment(deployed))
}

func successfulDeployment(deployed []*flowkitProject.Contract) string {
	var out bytes.Buffer
	okFaces := []string{"😎", "😲", "😱", "😜"}

	// build map of grouped contracts by account for easier output
	deployOut := make(map[string][]string)
	for _, deploy := range deployed {
		key := fmt.Sprintf("%s 0x%s", deploy.AccountName, deploy.AccountAddress.String())
		if deployOut[key] == nil {
			deployOut[key] = make([]string, 0)
		}
		deployOut[key] = append(
			deployOut[key],
			fmt.Sprintf("    |- %s  %s", output.Bold(output.Magenta(deploy.Name)), output.Italic(deploy.Location())),
		)
	}

	for account, contracts := range deployOut {
		out.WriteString(fmt.Sprintf("\n%s %s\n", okFaces[rand.Intn(len(okFaces))], output.Bold(account)))
		for _, contract := range contracts {
			out.WriteString(fmt.Sprintf("%s\n", contract))
		}
	}

	return out.String()
}

func failureDeployment(err error, contractPathNames map[string]string) string {
	var out bytes.Buffer

	// handle emulator not allowing overwriting contracts
	if strings.Contains(err.Error(), "cannot overwrite existing contract with name") {
		out.WriteString(output.ErrorEmoji() + output.Red(" Cannot overwrite existing contract, that means you are running the emulator without the --contract-removal flag.\n"))
		out.WriteString(output.TryEmoji() + " Please restart the emulator with the --contract-removal flag present as we are required to continuously update contracts as you work.")
	}

	// handle import path errors with helpful message
	importRegex := regexp.MustCompile(`import from (\w*) could not be found: (\w*), make sure import path is correct`)
	if importRegex.MatchString(err.Error()) {
		found := importRegex.FindAllStringSubmatch(err.Error(), -1)
		contractName := found[0][1]
		importName := found[0][2]
		contractPath := ""
		for p, n := range contractPathNames {
			if contractName == n {
				contractPath = p
			}
		}

		out.WriteString(output.ErrorEmoji() + output.Red(" Failed to sync your project, one of your imports is incorrect. Please fix the import detailed bellow.\n\n"))
		out.WriteString(fmt.Sprintf("Contract with error:  %s.cdc %s\n", contractName, contractPath))
		out.WriteString(fmt.Sprintf("Invalid import used:  import %s\n", importName))
		out.WriteString(fmt.Sprintf("Only valid imports:   %s\n", strings.Join(maps.Values(contractPathNames), ", ")))
		return out.String()
	}

	// handle cadence runtime errors
	var deployErr *services.ProjectDeploymentError
	if errors.As(err, &deployErr) {
		out.WriteString(output.ErrorEmoji() + output.Red(" Failed to deploy contracts due to runtime errors, this usually mean your code is incorrect. Check the detailed error bellow.\n\n"))

		for name, err := range deployErr.Contracts() {
			// remove transaction error as it confuses developer, the only important part is the actual code
			removeDeployOuput := regexp.MustCompile(`(?s)(failed deploying.*contracts\.add[^\n]*\n[^\n]*\n)`)
			out.WriteString(output.Bold(fmt.Sprintf("%s Errors:\n", name)))
			out.WriteString(output.Red(removeDeployOuput.ReplaceAllString(err.Error(), "")))
		}
		return out.String()
	}

	return err.Error()
}

func helpBanner() string {
	var out bytes.Buffer
	out.WriteString(output.Italic("The development environment will watch your Cadence files and automatically keep your project updated on the emulator.\n"))
	out.WriteString(output.Italic("Please add your contracts in the contracts folder, if you want to add a contract to a new account, create a folder\n"))
	out.WriteString(output.Italic("inside the contracts folder and we will automatically create an account for you and deploy everything inside that folder.\n\n"))
	return out.String()
}

func okBanner() string {
	return output.Bold(
		fmt.Sprintf("%s Project synced [%s]\n", output.Green("OK"), time.Now().Format("15:04:05")),
	)
}

func errorBanner() string {
	return output.Bold(
		fmt.Sprintf("%s Project error [%s]\n", output.Red("ERR"), time.Now().Format("15:04:05")),
	)
}

func clearScreen() {
	cmd := sysExec.Command("clear")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
