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
	"fmt"
	"os"

	"github.com/onflow/flow-cli/flowkit"

	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"

	"github.com/spf13/cobra"
)

type generateFlagsDef struct{}

var generateFlags = generateFlagsDef{}

var GenerateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "generate <contract | transaction | script>",
		Short:   "Generate new boilerplate files",
		Example: "flow generate HelloWorld",
		Args:    cobra.ArbitraryArgs,
		GroupID: "super",
	},
	Flags: &generateFlags,
	RunS:  generateNew,
}

func generateNew(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("invalid number of arguments")
	}

	templateType := args[0]
	name := args[1]
	filename := fmt.Sprintf("%s.cdc", name)
	var fileToWrite string

	if _, err := os.Stat(filename); err == nil {
		return nil, fmt.Errorf("file already exists: %s", filename)
	}

	contractTemplate := fmt.Sprintf(`
pub contract %s {
    // Contract details here
}`, name)

	scriptTemplate := `pub fun main() {
	// Script details here
}`

	transactionTemplate := `transaction() {
    prepare() {}

    execute {}
}`

	switch templateType {
	case "contract":
		fileToWrite = contractTemplate
	case "script":
		fileToWrite = scriptTemplate
	case "transaction":
		fileToWrite = transactionTemplate
	default:
		return nil, fmt.Errorf("invalid template type: %s", templateType)
	}

	err = os.WriteFile(filename, []byte(fileToWrite), 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing file: %w", err)
	}

	logger.Info(fmt.Sprintf("Generated new contract: %s at %s", name, filename))

	return nil, err
}
