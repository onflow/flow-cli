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

package accounts

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type flagsList struct {
	Sort bool `default:"false" flag:"sort" info:"Sort account names"`
}

var listFlags = flagsList{
	Sort: false,
}

type ListResult struct {
	accounts []string
}

var ListCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "list",
		Short:   "Lists all accounts",
		Example: "flow accounts list",
		Args:    cobra.ExactArgs(0),
	},
	Flags: &listFlags,
	Run:   list,
}

func list(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	accounts := services.Accounts.List()

	return &ListResult{accounts}, nil
}

func (r *ListResult) JSON() interface{} {
	result := make(map[string]interface{})
	result["accounts"] = r.accounts

	return result
}

func (r *ListResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	if listFlags.Sort {
		sort.Strings(r.accounts)
	}
	accounts := strings.Join(r.accounts[:], ",")
	fmt.Fprintf(writer, "%v", accounts)

	writer.Flush()
	return b.String()
}

func (r *ListResult) Oneliner() string {
	return r.String()
}
