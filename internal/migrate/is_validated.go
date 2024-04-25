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
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/logrusorgru/aurora/v4"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

//go:generate mockery --name GitHubRepositoriesService --output ./mocks --case underscore
type GitHubRepositoriesService interface {
	GetContents(ctx context.Context, owner string, repo string, path string, opt *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
	DownloadContents(ctx context.Context, owner string, repo string, filepath string, opt *github.RepositoryContentGetOptions) (io.ReadCloser, error)
}

type validator struct {
	repoService GitHubRepositoriesService
	state       *flowkit.State
	logger      output.Logger
	network     config.Network
}

type contractUpdateStatus struct {
	Kind           string `json:"kind,omitempty"`
	AccountAddress string `json:"account_address"`
	ContractName   string `json:"contract_name"`
	Error          string `json:"error,omitempty"`
}

type validationResult struct {
	Timestamp time.Time
	Status    contractUpdateStatus
	Network   string
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

const moreInformationMessage = "For more information, please find the latest full migration report on GitHub (https://github.com/onflow/cadence/tree/master/migrations_data).\n\nNew reports are generated after each weekly emulated migration and your contract's status may change, so please actively monitor this status and stay tuned for the latest announcements until the migration deadline."
const contractUpdateFailureKind = "contract-update-failure"

func isValidated(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	repoService := github.NewClient(nil).Repositories
	v := newValidator(repoService, flow.Network(), state, logger)

	contractName := args[0]
	return v.validate(contractName)
}

func newValidator(repoService GitHubRepositoriesService, network config.Network, state *flowkit.State, logger output.Logger) *validator {
	return &validator{
		repoService: repoService,
		state:       state,
		logger:      logger,
		network:     network,
	}
}

func (v *validator) validate(contractName string) (validationResult, error) {
	err := checkNetwork(v.network)
	if err != nil {
		return validationResult{}, err
	}

	v.logger.StartProgress("Checking if contract has been validated")
	defer v.logger.StopProgress()

	addr, err := getAddressByContractName(v.state, contractName, v.network)
	if err != nil {
		return validationResult{}, err
	}

	status, timestamp, err := v.getContractValidationStatus(
		v.network,
		addr.HexWithPrefix(),
		contractName,
	)
	if err != nil {
		// Append more information message to the error
		// this way we can ensure that if, for whatever reason, we fail to fetch the report
		// the user will still understand that they can find the report on GitHub
		return validationResult{}, fmt.Errorf("%w\n\n%s%s", err, moreInformationMessage, "\n")
	}

	return validationResult{
		Timestamp: *timestamp,
		Status:    status,
		Network:   v.network.Name,
	}, nil
}

func (v *validator) getContractValidationStatus(network config.Network, address string, contractName string) (contractUpdateStatus, *time.Time, error) {
	// Get last migration report
	report, timestamp, err := v.getLatestMigrationReport(network)
	if err != nil {
		return contractUpdateStatus{}, nil, err
	}

	// Get all the contract statuses from the report
	statuses, err := v.fetchAndParseReport(report.GetPath())
	if err != nil {
		return contractUpdateStatus{}, nil, err
	}

	// Get the validation result related to the contract
	var status *contractUpdateStatus
	for _, s := range statuses {
		if s.ContractName == contractName && s.AccountAddress == address {
			status = &s
			break
		}
	}

	// Throw error if contract was not part of the last migration
	if status == nil {
		builder := strings.Builder{}
		builder.WriteString("the contract does not appear to have been a part of any emulated migrations yet, please ensure that it has been staged & wait for the next emulated migration (last migration report was at ")
		builder.WriteString(timestamp.Format(time.RFC3339))
		builder.WriteString(")\n\n")

		builder.WriteString(" - Account: ")
		builder.WriteString(address)
		builder.WriteString("\n - Contract: ")
		builder.WriteString(contractName)
		builder.WriteString("\n - Network: ")
		builder.WriteString(network.Name)

		return contractUpdateStatus{}, nil, fmt.Errorf(builder.String())
	}

	return *status, timestamp, nil
}

func (v *validator) getLatestMigrationReport(network config.Network) (*github.RepositoryContent, *time.Time, error) {
	// Get the content of the migration reports folder
	_, folderContent, _, err := v.repoService.GetContents(
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

			// Try to extract the time from the filename
			networkStr, t, err := extractInfoFromFilename(contentPath)
			if err != nil {
				// Ignore files that don't match the expected format
				// Or have any another error while parsing
				continue
			}

			// Ignore reports from other networks
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
		return nil, nil, fmt.Errorf("no emulated migration reports found for network `%s` within the remote repository - have any migrations been run yet for this network?", network.Name)
	}

	return latestReport, latestReportTime, nil
}

func (v *validator) fetchAndParseReport(reportPath string) ([]contractUpdateStatus, error) {
	// Get the content of the latest report
	rc, err := v.repoService.DownloadContents(
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

	expr := regexp.MustCompile(`^staged-contracts-report.*(\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z)-([a-z]+).json$`)
	regexpMatches := expr.FindStringSubmatch(fileName)
	if regexpMatches == nil {
		return "", nil, fmt.Errorf("filename does not match the expected format")
	}

	// Extract the timestamp
	timestampStr := regexpMatches[1]
	timestamp, err := time.Parse("2006-01-02T15-04-05Z", timestampStr)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse timestamp from filename")
	}

	// Extract the network
	network := regexpMatches[2]

	return network, &timestamp, nil
}

func (s contractUpdateStatus) IsFailure() bool {
	// Just in case there are failures without an error message in the future
	// we will also check the kind of the status
	return s.Error != "" || s.Kind == contractUpdateFailureKind
}

func (v validationResult) String() string {
	status := v.Status

	builder := strings.Builder{}
	builder.WriteString("Last emulated migration report was created at ")
	builder.WriteString(v.Timestamp.Format(time.RFC3339))
	builder.WriteString("\n\n")

	statusBuilder := strings.Builder{}
	emoji := "✅ "
	statusColor := aurora.Green
	if status.IsFailure() {
		emoji = "❌ "
		statusColor = aurora.Red
	}

	statusBuilder.WriteString(util.PrintEmoji(emoji))
	statusBuilder.WriteString("The contract has ")

	if status.IsFailure() {
		statusBuilder.WriteString("FAILED")
	} else {
		statusBuilder.WriteString("PASSED")
	}
	statusBuilder.WriteString(" the last emulated migration")

	statusBuilder.WriteString("\n\n - Account: ")
	statusBuilder.WriteString(status.AccountAddress)
	statusBuilder.WriteString("\n - Contract: ")
	statusBuilder.WriteString(status.ContractName)
	statusBuilder.WriteString("\n - Network: ")
	statusBuilder.WriteString(v.Network)
	statusBuilder.WriteString("\n\n")

	// Write colored status
	builder.WriteString(statusColor(statusBuilder.String()).String())

	if status.Error != "" {
		builder.WriteString(status.Error)
		builder.WriteString("\n")
	}

	if status.IsFailure() {
		builder.WriteString(aurora.Red(">> Please review the error and re-stage the contract to resolve these issues if necessary").String())
		builder.WriteString("\n\n")
	}

	builder.WriteString(moreInformationMessage)
	return builder.String()
}

func (v validationResult) JSON() interface{} {
	return v
}

func (v validationResult) Oneliner() string {
	if v.Status.IsFailure() {
		return util.MessageWithEmojiPrefix("❌", "FAILED")
	}
	return util.MessageWithEmojiPrefix("✅", "FAIlED")
}
