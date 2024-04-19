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
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

//go:generate mockery --name GitHubRepositoriesService --output ./mocks --case underscore
type GitHubRepositoriesService interface {
	GetContents(ctx context.Context, owner string, repo string, path string, opt *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
}

type contractUpdateStatus struct {
	AccountAddress string `json:"account_address"`
	ContractName   string `json:"contract_name"`
	Error          string `json:"error"`
}

type validationResult struct {
	Timestamp time.Time
	Status    contractUpdateStatus
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
	RunS:  isValidated(nil),
}

const (
	repoOwner = "onflow"
	repoName  = "cadence"
	repoPath  = "migrations_data"
	repoRef   = "master"
)

func isValidated(repoService GitHubRepositoriesService) func(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	return func(
		args []string,
		globalFlags command.GlobalFlags,
		logger output.Logger,
		flow flowkit.Services,
		state *flowkit.State,
	) (command.Result, error) {
		if repoService == nil {
			repoService = github.NewClient(nil).Repositories
		}

		logger.StartProgress("Checking if contract has been validated")
		defer logger.StopProgress()

		contractName := args[0]
		status, timestamp, err := getContractValidationStatus(contractName, flow.Network(), state, repoService)
		if err != nil {
			return nil, err
		}

		return validationResult{
			Timestamp: *timestamp,
			Status:    *status,
		}, nil
	}
}

func getContractValidationStatus(contractName string, network config.Network, state *flowkit.State, repoService GitHubRepositoriesService) (*contractUpdateStatus, *time.Time, error) {
	addr, err := getAddressByContractName(state, contractName, network)
	if err != nil {
		return nil, nil, err
	}
	addrHex := addr.HexWithPrefix()

	report, timestamp, err := getLatestMigrationReport(repoService)
	if err != nil {
		return nil, nil, err
	}

	statuses, err := fetchAndParseReport(repoService, report.GetPath())
	if err != nil {
		return nil, nil, err
	}

	// get the validation result
	var status *contractUpdateStatus
	for _, s := range statuses {
		if s.ContractName == contractName && s.AccountAddress == addrHex {
			status = &s
			break
		}
	}

	if status == nil {
		return nil, nil, fmt.Errorf("the contract %s has not been part of any emulated migrations yet, please ensure it is staged & wait for the next emulated migration (last migration: %s)", contractName, timestamp.Format(time.RFC3339))
	}

	return status, timestamp, nil
}

func getLatestMigrationReport(repoService GitHubRepositoriesService) (*github.RepositoryContent, *time.Time, error) {
	// get tree folder containing the reports
	_, folderContent, _, err := repoService.GetContents(
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
			if latestReportTime == nil || t.After(*latestReportTime) {
				latestReport = content
				latestReportTime = t
			}
		}
	}

	if latestReport == nil {
		return nil, nil, fmt.Errorf("no reports found")
	}

	return latestReport, latestReportTime, nil
}

func fetchAndParseReport(repoService GitHubRepositoriesService, reportPath string) ([]contractUpdateStatus, error) {
	// Get the content of the latest report
	reportContent, _, _, err := repoService.GetContents(
		context.Background(),
		repoOwner,
		repoName,
		reportPath,
		&github.RepositoryContentGetOptions{
			Ref: repoRef,
		},
	)
	if err != nil {
		return nil, err
	}

	// Parse the report
	var statuses []contractUpdateStatus
	reportStr, err := reportContent.GetContent()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(reportStr), &statuses)
	if err != nil {
		return nil, err
	}

	return statuses, nil
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
	status := v.Status

	builder := strings.Builder{}
	builder.WriteString("Last emulated migration occured at: ")
	builder.WriteString(v.Timestamp.Format(time.RFC3339))
	builder.WriteString("\n\n")

	builder.WriteString("The contract, ")
	builder.WriteString(status.ContractName)
	builder.WriteString(", has ")

	if status.Error == "" {
		builder.WriteString("PASSED the last emulated migration")
	} else {
		builder.WriteString("FAILED the last emulated migration")
	}

	builder.WriteString("\n\n")

	if status.Error != "" {
		builder.WriteString("Error: ")
		builder.WriteString(status.Error)
	}
	return builder.String()
}

func (v validationResult) JSON() interface{} {
	return v
}

func (v validationResult) Oneliner() string {
	if v.Status.Error == "" {
		return "PASSED"
	}
	return "FAILED"
}
