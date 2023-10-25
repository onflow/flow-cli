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
	"strings"

	"github.com/onflow/flow-cli/flowkit"

	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"

	"github.com/spf13/cobra"
)

type generateFlagsDef struct{}

var generateFlags = generateFlagsDef{}

var GenerateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "generate <contract | transaction | script> <name>",
		Short:   "Generate new boilerplate files",
		Example: "flow generate contract HelloWorld",
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
	var filename string

	// Don't add .cdc extension if it's already there
	if strings.HasSuffix(name, ".cdc") {
		filename = name
	} else {
		filename = fmt.Sprintf("%s.cdc", name)
	}

	var fileToWrite string
	var basePath string

	switch templateType {
	case "contract":
		basePath = "cadence/contracts"
		fileToWrite = fmt.Sprintf(`
pub contract %s {
    // Contract details here
}`, name)
	case "script":
		basePath = "cadence/scripts"
		fileToWrite = `pub fun main() {
    // Script details here
}`
	case "transaction":
		basePath = "cadence/transactions"
		fileToWrite = `transaction() {
    prepare() {}

    execute {}
}`
	default:
		return nil, fmt.Errorf("invalid template type: %s", templateType)
	}

	filenameWithBasePath := filepath.Join(basePath, filename)

	if _, err := os.Stat(filenameWithBasePath); err == nil {
		return nil, fmt.Errorf("file already exists: %s", filenameWithBasePath)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("error creating directories: %w", err)
	}

	err = os.WriteFile(filenameWithBasePath, []byte(fileToWrite), 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing file: %w", err)
	}

	logger.Info(fmt.Sprintf("Generated new contract: %s at %s", name, filenameWithBasePath))

	return nil, err
}
