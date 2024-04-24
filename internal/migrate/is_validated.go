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
	"io"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/logrusorgru/aurora/v4"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

//go:generate mockery --name GitHubRepositoriesService --output ./mocks --case underscore
type GitHubRepositoriesService interface {
	GetContents(ctx context.Context, owner string, repo string, path string, opt *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
	DownloadContents(ctx context.Context, owner string, repo string, filepath string, opt *github.RepositoryContentGetOptions) (io.ReadCloser, error)
}

type contractUpdateStatus struct {
	AccountAddress string `json:"account_address"`
	ContractName   string `json:"contract_name"`
	Error          string `json:"error,omitempty"`
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
		err := checkNetwork(flow.Network())
		if err != nil {
			return nil, err
		}

		if repoService == nil {
			repoService = github.NewClient(nil).Repositories
		}

		logger.StartProgress("Checking if contract has been validated")
		defer logger.StopProgress()

		contractName := args[0]
		addr, err := getAddressByContractName(state, contractName, flow.Network())
		if err != nil {
			return nil, err
		}

		status, timestamp, err := getContractValidationStatus(
			flow.Network(),
			addr.HexWithPrefix(),
			contractName,
			state,
			repoService,
			logger,
		)
		if err != nil {
			return nil, err
		}

		return validationResult{
			Timestamp: *timestamp,
			Status:    *status,
		}, nil
	}
}

func getContractValidationStatus(network config.Network, address string, contractName string, state *flowkit.State, repoService GitHubRepositoriesService, logger output.Logger) (*contractUpdateStatus, *time.Time, error) {
	// Get last migration report
	report, timestamp, err := getLatestMigrationReport(network, repoService, logger)
	if err != nil {
		return nil, nil, err
	}

	// Get all the contract statuses from the report
	statuses, err := fetchAndParseReport(repoService, report.GetPath())
	if err != nil {
		return nil, nil, err
	}

	// Gett the validation result related to the contract
	var status *contractUpdateStatus
	for _, s := range statuses {
		if s.ContractName == contractName && s.AccountAddress == address {
			status = &s
			break
		}
	}

	// Throw error if contract was not part of the last migration
	if status == nil {
		return nil, nil, fmt.Errorf("the contract %s has not been part of any emulated migrations yet, please ensure it is staged & wait for the next emulated migration (last migration report was at %s)", contractName, timestamp.Format(time.RFC3339))
	}

	return status, timestamp, nil
}

func getLatestMigrationReport(network config.Network, repoService GitHubRepositoriesService, logger output.Logger) (*github.RepositoryContent, *time.Time, error) {
	// Get the content of the migration reports folder
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

			// Extract the time from the filename
			networkStr, t, err := extractInfoFromFilename(contentPath)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to extract report information from filename, file appears to be in an unexpected format: %s", contentPath))
				continue
			}

			if networkStr != strings.ToLower(network.Name) {
				continue
			}

			// Check if this is the latest report
			if latestReportTime == nil || t.After(*latestReportTime) {
				latestReport = content
				latestReportTime = t
			}
		}
	}

	if latestReport == nil {
		return nil, nil, fmt.Errorf("no reports found on the repository, have any migrations been run yet?")
	}

	return latestReport, latestReportTime, nil
}

func fetchAndParseReport(repoService GitHubRepositoriesService, reportPath string) ([]contractUpdateStatus, error) {
	// Get the content of the latest report
	rc, err := repoService.DownloadContents(
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
	defer rc.Close()

	// Read the report content
	reportContent, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	// Parse the report
	var statuses []contractUpdateStatus
	err = json.Unmarshal(reportContent, &statuses)
	if err != nil {
		return nil, err
	}

	return statuses, nil
}

func extractInfoFromFilename(filename string) (string, *time.Time, error) {
	// Extracts the timestamp from the filename in the format: migrations_data/raw/XXXXXX-MM-DD-YYYY-<network>-XXXXXX.json
	fileName := path.Base(filename)
	fileNameWithoutExt := strings.TrimSuffix(fileName, path.Ext(fileName))
	splitFileName := strings.Split(fileNameWithoutExt, "-")

	fmt.Println(splitFileName)
	if len(splitFileName) < 4 {
		return "", nil, fmt.Errorf("filename is not in the expected format")
	}

	// Data type is the first elements
	if splitFileName[0] != "staged" || splitFileName[1] != "contracts" || splitFileName[2] != "report" {
		return "", nil, fmt.Errorf("filename is not in the expected format")
	}

	// Last elements excluding very last one are the timestamp
	dateTimeSplit := splitFileName[len(splitFileName)-4 : len(splitFileName)-1]

	// Extract the timestamp
	timestampStr := strings.Join(dateTimeSplit, "-")
	timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse timestamp from filename")
	}

	// Extract the network
	network := strings.ToLower(splitFileName[len(splitFileName)-1])

	return network, &timestamp, nil
}

func (v validationResult) String() string {
	status := v.Status

	builder := strings.Builder{}
	builder.WriteString("Last emulated migration report was created at ")
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
		builder.WriteString(status.Error)
		builder.WriteString("\n")
		builder.WriteString(aurora.Red("Please review the error and re-stage the contract to resolve these issues if necessary\n").String())
	}

	builder.WriteString("\n")
	builder.WriteString("For more information, please find the latest full migration report on GitHub: https://github.com/onflow/cadence/tree/master/migrations_data\n")

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
