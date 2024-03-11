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

package collections

import (
	"context"
	"fmt"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsCollections struct{}

var collectionFlags = flagsCollections{}

var getCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get <collection_id>",
		Short:   "Get collection info",
		Example: "flow collections get 270d...9c31e",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &collectionFlags,
	Run:   get,
}

func get(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	id := flowsdk.HexToID(args[0])

	logger.StartProgress(fmt.Sprintf("Loading collection %s", id))
	defer logger.StopProgress()

	collection, err := flow.GetCollection(context.Background(), id)
	if err != nil {
		return nil, err
	}

	return &collectionResult{collection}, nil
}
