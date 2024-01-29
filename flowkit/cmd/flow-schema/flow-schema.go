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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	configJson "github.com/onflow/flow-cli/flowkit/v2/config/json"
)

func main() {
	var verify bool
	flag.BoolVar(&verify, "verify", false, "Verify the schema")

	flag.Parse()
	path := flag.Arg(0)

	if path == "" {
		fmt.Println("Path is required")
		os.Exit(1)
	}

	schema := configJson.GenerateSchema()
	json, err := json.MarshalIndent(schema, "", "  ")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if verify {
		fileContents, err := os.ReadFile(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if normalizeNewlines(string(fileContents)) != normalizeNewlines(string(json)) {
			fmt.Println("Schema is out of date - have you run `make generate-schema`?")
			os.Exit(1)
		}
	} else {
		os.WriteFile(path, json, 0644)
	}
}

func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
