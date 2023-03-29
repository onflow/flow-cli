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

package emulator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/spf13/cobra"
)

type SnapshotFlag struct {
}

var snapshotFlag = SnapshotFlag{}

var SnapshotCmd = &command.Command{
	Cmd: &cobra.Command{
		Use:     "snapshot <snapshotName>",
		Short:   "Save/Load a emulator snapshot",
		Example: "flow emulator snapshot testSnapshot",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &snapshotFlag,
	Run:   snapshot,
}

type SnapShotResult struct {
	Name    string `json:"context"`
	BlockID string `json:"blockId"`
	Height  uint64 `json:"height"`
	Result  string `json:"result,omitempty"`
}

func (r *SnapShotResult) JSON() interface{} {
	result := make(map[string]interface{})
	result["name"] = r.Name
	result["blockID"] = r.BlockID
	result["height"] = r.Height
	if r.Result != "" {
		result["result"] = r.Result
	}

	return result
}

func (r *SnapShotResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)
	if r.Result != "" {
		_, _ = fmt.Fprintf(writer, "%s\n", r.Result)
	}
	_, _ = fmt.Fprintf(writer, "Name\t%s\n", r.Name)
	_, _ = fmt.Fprintf(writer, "Block ID\t%s\n", r.BlockID)
	_, _ = fmt.Fprintf(writer, "Height\t%d", r.Height)

	_ = writer.Flush()
	return b.String()
}

func (r *SnapShotResult) Oneliner() string {
	return fmt.Sprintf("%s : %s (%d) %s", r.Name, r.BlockID, r.Height, r.Result)
}

func listSnapshot() (result []string, err error) {

	snapshots, err := http.Get("http://localhost:8080/emulator/snapshots")
	if err != nil {
		return []string{}, err
	}

	if snapshots.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("unable to list snapshots on the emulator")
	}

	body, err := io.ReadAll(snapshots.Body)
	if err != nil {
		return []string{}, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return []string{}, err
	}

	return result, nil
}

func snapshot(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
) (command.Result, error) {

	name := args[0]

	snapshots, err := listSnapshot()
	if err != nil {
		return nil, err
	}

	exists := false
	for _, snapshotName := range snapshots {
		if name == snapshotName {
			exists = true
			break
		}
	}

	if exists {
		//load snapshot
		requestBody := bytes.NewBufferString("")
		request, err := http.NewRequest("PUT",
			fmt.Sprintf("http://localhost:8080/emulator/snapshots/%s", name),
			requestBody)

		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(request)

		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unable to create snapshot on the emulator")
		}

		var result SnapShotResult
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		result.Result = "Snapshot loaded"
		return &result, nil
	} else {
		//save snapshot
		requestBody := bytes.NewBufferString(fmt.Sprintf("name=%s", name))
		resp, err := http.Post("http://localhost:8080/emulator/snapshots",
			"application/x-www-form-urlencoded",
			requestBody)

		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unable to create snapshot on the emulator")
		}

		var result SnapShotResult
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		result.Result = "Snapshot created"
		return &result, nil
	}

}
