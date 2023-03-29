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
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"strings"

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
		Use:     "snapshot <create|load|list> [snapshotName]",
		Short:   "Create/Load/List emulator snapshots",
		Example: "flow emulator snapshot create testSnapshot",
		Args:    cobra.RangeArgs(1, 2),
	},
	Flags: &snapshotFlag,
	Run:   snapshot,
}

type SnapshotList struct {
	Snapshots []string
}

func (s *SnapshotList) JSON() interface{} {
	return s.Snapshots
}

func (s *SnapshotList) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)
	_, _ = fmt.Fprintf(writer, "Snapshots:\n")
	for _, snapshotName := range s.Snapshots {
		_, _ = fmt.Fprintf(writer, "\t%s\n", snapshotName)
	}
	_ = writer.Flush()
	return b.String()
}

func (s *SnapshotList) Oneliner() string {
	return strings.Join(s.Snapshots, ",")
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

const SnapshotEndpoint = "http://localhost:8080/emulator/snapshots"

func makeRequest(r *http.Request, v any) error {
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("emulator snapshot request error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("emulator snapshot request error: status_code=%d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}

	return nil
}

func requestListSnapshots() (*http.Request, error) {
	request, err := http.NewRequest("GET", SnapshotEndpoint, nil)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func requestCreateSnapshot(name string) (*http.Request, error) {
	requestBody := bytes.NewBufferString(fmt.Sprintf("name=%s", name))
	request, err := http.NewRequest("POST", SnapshotEndpoint, requestBody)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return nil, err
	}
	return request, nil
}

func requestLoadSnapshot(name string) (*http.Request, error) {
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", SnapshotEndpoint, name), nil)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func listSnapshot() (result []string, err error) {
	req, err := requestListSnapshots()
	if err != nil {
		return []string{}, err
	}

	err = makeRequest(req, &result)
	if err != nil {
		return []string{}, err
	}

	return result, nil
}

type SnapshotCommand string

const (
	SnapshotCommandList   SnapshotCommand = "list"
	SnapshotCommandCreate SnapshotCommand = "create"
	SnapshotCommandLoad   SnapshotCommand = "load"
)

func snapshot(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
) (command.Result, error) {

	subCommand := args[0]

	snapshots, err := listSnapshot()
	if err != nil {
		return nil, err
	}

	switch SnapshotCommand(subCommand) {
	case SnapshotCommandList:
		return &SnapshotList{Snapshots: snapshots}, nil

	case SnapshotCommandCreate:
		if len(args) < 2 {
			return nil, fmt.Errorf("snapshot create command requires name argument")
		}
		name := args[1]
		exists := slices.Contains(snapshots, name)
		if exists {
			return nil, fmt.Errorf("snapshot '%s' already exists", name)
		}

		var result SnapShotResult
		req, err := requestCreateSnapshot(name)
		if err != nil {
			return nil, err
		}
		err = makeRequest(req, &result)
		if err != nil {
			return nil, err
		}
		result.Result = "Snapshot created"
		return &result, nil

	case SnapshotCommandLoad:
		if len(args) < 2 {
			return nil, fmt.Errorf("snapshot load command requires name argument")
		}
		name := args[1]
		exists := slices.Contains(snapshots, name)
		if !exists {
			return nil, fmt.Errorf("snapshot '%s' does not exist", name)
		}

		var result SnapShotResult
		req, err := requestLoadSnapshot(name)
		if err != nil {
			return nil, err
		}
		err = makeRequest(req, &result)
		if err != nil {
			return nil, err
		}
		result.Result = "Snapshot loaded"
		return &result, nil

	default:
		return nil, fmt.Errorf("invalid snapshot command: valid commands are: 'list', 'create', 'load'")
	}

}
