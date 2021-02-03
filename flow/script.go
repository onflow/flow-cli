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

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func ExecuteScript(host string, script []byte, arguments ...cadence.Value) {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}
	value, err := flowClient.ExecuteScriptAtLatestBlock(ctx, script, arguments)
	if err != nil {
		Exitf(1, "Failed to submit executable script: %s", err)
	}
	b, err := jsoncdc.Encode(value)
	if err != nil {
		Exitf(1, "Failed to decode cadence value: %s", err)
	}
	// TODO: Consider using printField function, once we're adapted the function to print more cadence types
	// For now, just pretty print the JSON return object, at least
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, b, Indent, Indent)
	if err != nil {
		Exitf(1, "Failed to print cadence value: %s", err)
	}
	PrintIndent(1)
	fmt.Println(prettyJSON.String())
}