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

package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

type contractUpdateStatus struct {
	AccountAddress string `json:"account_address"`
	ContractName   string `json:"contract_name"`
	Error          string `json:"error"`
}

type validationResult struct {
	AccountAddress string
	ContractName   string
	WasChecked     bool
	Error          string
	Timestamp      time.Time
}

var isValidatedflags struct{}

var IsValidatedCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "is-validated <CONTRACT_NAME>",
		Short:   "checks to see if the contract has passed the last emulated migration",
		Example: `flow migrate is-validated HelloWorld`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &isValidatedflags,
	RunS:  isValidated,
}

const (
	repoOwner = "onflow"
	repoName  = "cadence"
	repoPath  = "migrations_data"
	repoRef   = "master"
)

func isValidated(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	contractName := args[0]

	addr, err := getAddressByContractName(state, contractName, flow.Network())
	if err != nil {
		return nil, err
	}
	addressHex := addr.HexWithPrefix()

	logger.StartProgress("Checking if contract has been validated")
	defer logger.StopProgress()

	statuses, timestamp, err := getLatestMigrationReport()
	if err != nil {
		return nil, err
	}

	// get the validation result
	var status *contractUpdateStatus
	for _, s := range statuses {
		if s.ContractName == contractName && s.AccountAddress == addressHex {
			status = &s
			break
		}
	}

	logger.StopProgress()

	errorMessage := ""
	if status != nil {
		errorMessage = status.Error
	}

	return validationResult{
		AccountAddress: status.AccountAddress,
		ContractName:   status.ContractName,
		Timestamp:      *timestamp,
		Error:          errorMessage,
		WasChecked:     status != nil,
	}, nil
}

func getLatestMigrationReport() ([]contractUpdateStatus, *time.Time, error) {
	gitClient := github.NewClient(nil)

	// get tree folder containing the reports
	_, folderContent, _, err := gitClient.Repositories.GetContents(
		context.Background(),
		repoOwner,
		repoName,
		repoPath,
		&github.RepositoryContentGetOptions{
			Ref: repoRef,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	// Find the latest report file
	var latestReport *github.RepositoryContent
	var latestReportTime *time.Time
	for _, content := range folderContent {
		if content.Type != nil && *content.Type == "file" {
			contentPath := content.GetPath()
			if path.Ext(contentPath) != ".json" {
				continue
			}

			// extract the time from the filename
			t, err := extractTimeFromFilename(contentPath)
			if err != nil {
				return nil, nil, err
			}

			// check if this is the latest report
			if t.After(*latestReportTime) {
				latestReport = content
				latestReportTime = t
			}
		}
	}

	if latestReport == nil {
		return nil, nil, fmt.Errorf("no reports found")
	}

	// Get the content of the latest report
	reportContent, _, _, err := gitClient.Repositories.GetContents(
		context.Background(),
		repoOwner,
		repoName,
		latestReport.GetPath(),
		&github.RepositoryContentGetOptions{
			Ref: repoRef,
		},
	)

	if err != nil {
		return nil, nil, err
	}

	// Parse the report
	var statuses []contractUpdateStatus
	reportStr, err := reportContent.GetContent()
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal([]byte(reportStr), &statuses)
	if err != nil {
		return nil, nil, err
	}

	return statuses, latestReportTime, nil
}

func extractTimeFromFilename(filename string) (*time.Time, error) {
	var splitFileName []string
	fileName := path.Base(filename)
	fileNameWithoutExt := strings.TrimSuffix(fileName, path.Ext(fileName))
	splitFileName = strings.Split(fileNameWithoutExt, "-")
	timestampStr := splitFileName[len(splitFileName)-1]
	unixTimestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, err
	}

	t := time.Unix(unixTimestamp, 0)
	return &t, nil
}

func (v validationResult) String() string {
	builder := strings.Builder{}
	builder.WriteString("Last emulated migration occured at: ")
	builder.WriteString(v.Timestamp.Format(time.RFC3339))
	builder.WriteString("\n\n")

	builder.WriteString("The contract, ")
	builder.WriteString(v.ContractName)
	builder.WriteString(", has ")
	if v.WasChecked {
		if v.Error == "" {
			builder.WriteString("PASSED the last emulated migration")
		} else {
			builder.WriteString("FAILED the last emulated migration")
		}
	} else {
		builder.WriteString("not been part of any emulated migrations")
	}
	builder.WriteString("\n\n")

	if v.Error != "" {
		builder.WriteString("Error: ")
		builder.WriteString(v.Error)
	}
	return builder.String()
}

func (v validationResult) JSON() interface{} {
	return v
}

func (v validationResult) Oneliner() string {
	if v.WasChecked {
		if v.Error == "" {
			return "PASSED"
		}
		return "FAILED"
	} else {
		return "NOT CHECKED"
	}
}
