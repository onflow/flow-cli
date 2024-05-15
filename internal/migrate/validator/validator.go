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

package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/google/go-github/github"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
)

//go:generate mockery --name GitHubRepositoriesService --output ./mocks --case underscore
type GitHubRepositoriesService interface {
	GetContents(ctx context.Context, owner string, repo string, path string, opt *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
	DownloadContents(ctx context.Context, owner string, repo string, filepath string, opt *github.RepositoryContentGetOptions) (io.ReadCloser, error)
}

const contractUpdateFailureKind = "contract-update-failure"

const (
	repoOwner = "onflow"
	repoName  = "cadence"
	repoPath  = "migrations_data"
	repoRef   = "master"
)

type missingContractError struct {
	MissingContracts []struct {
		ContractName string
		Address      string
		Network      string
	}
	LastMigrationTime *time.Time
}

func (m missingContractError) Error() string {
	builder := strings.Builder{}
	builder.WriteString("some contracts do not appear to have been a part of any emulated migrations yet, please ensure that it has been staged & wait for the next emulated migration (last migration report was at ")
	builder.WriteString(m.LastMigrationTime.Format(time.RFC3339))
	builder.WriteString(")\n\n")

	for _, contract := range m.MissingContracts {
		builder.WriteString(" - Account: ")
		builder.WriteString(contract.Address)
		builder.WriteString("\n - Contract: ")
		builder.WriteString(contract.ContractName)
		builder.WriteString("\n - Network: ")
		builder.WriteString(contract.Network)
		builder.WriteString("\n\n")
	}

	return builder.String()
}

type validator struct {
	repoService GitHubRepositoriesService
	state       *flowkit.State
	logger      output.Logger
	network     config.Network
}

type ContractUpdateStatus struct {
	Kind           string `json:"kind,omitempty"`
	AccountAddress string `json:"account_address"`
	ContractName   string `json:"contract_name"`
	Error          string `json:"error,omitempty"`
}

func (s ContractUpdateStatus) IsFailure() bool {
	// Just in case there are failures without an error message in the future
	// we will also check the kind of the status
	return s.Error != "" || s.Kind == contractUpdateFailureKind
}

func NewValidator(repoService GitHubRepositoriesService, network config.Network, state *flowkit.State, logger output.Logger) *validator {
	return &validator{
		repoService: repoService,
		state:       state,
		logger:      logger,
		network:     network,
	}
}

func (v *validator) GetContractStatuses() ([]ContractUpdateStatus, error) {
	if v.state == nil || v.state.Contracts() == nil {
		return nil, nil
	}

	var contractNames []string
	for _, c := range *v.state.Contracts() {
		contractNames = append(contractNames, c.Name)
	}

	statuses, _, err := v.getContractUpdateStatuses(contractNames...)
	if err != nil {
		return nil, err
	}

	return statuses, err
}

func (v *validator) getContractUpdateStatuses(contractNames ...string) ([]ContractUpdateStatus, *time.Time, error) {
	var contractUpdateStatuses []ContractUpdateStatus
	err := util.CheckNetwork(v.network)
	if err != nil {
		return nil, nil, err
	}

	v.logger.StartProgress("Checking if contracts has been validated...")
	defer v.logger.StopProgress()

	addressToContractName := make(map[string]string)
	for _, contractName := range contractNames {
		addr, err := util.GetAddressByContractName(v.state, contractName, v.network)
		if err != nil {
			return nil, nil, err
		}
		addressToContractName[addr.HexWithPrefix()] = contractName
	}

	// Get last migration report
	report, ts, err := v.getLatestMigrationReport(v.network)
	if err != nil {
		return nil, nil, err
	}

	// Get all the contract statuses from the report
	statuses, err := v.fetchAndParseReport(report.GetPath())
	if err != nil {
		return nil, nil, err
	}

	// Get the validation result related to the contract
	var foundAddresses []string
	for _, s := range statuses {
		if addressToContractName[s.AccountAddress] == s.ContractName {
			contractUpdateStatuses = append(contractUpdateStatuses, s)
			foundAddresses = append(foundAddresses, s.AccountAddress)
		}
	}

	for addr, contractName := range addressToContractName {
		var missingContractErr missingContractError
		if !slices.Contains(foundAddresses, addr) {
			missingContractErr.MissingContracts = append(missingContractErr.MissingContracts, struct {
				ContractName string
				Address      string
				Network      string
			}{
				ContractName: contractName,
				Address:      addr,
				Network:      v.network.Name,
			})
			missingContractErr.LastMigrationTime = ts
		}

		if len(missingContractErr.MissingContracts) > 0 {
			return nil, nil, missingContractErr
		}
	}

	return contractUpdateStatuses, ts, nil
}

func (v *validator) getContractValidationStatus(contractName string) (ContractUpdateStatus, *time.Time, error) {
	status, ts, err := v.getContractUpdateStatuses(contractName)
	if err != nil {
		return ContractUpdateStatus{}, nil, err
	}

	if len(status) != 1 {
		return ContractUpdateStatus{}, nil, fmt.Errorf("failed to find contract in last migration report")
	}

	return status[0], ts, nil

}

func (v *validator) Validate(contractName string) (ContractUpdateStatus, *time.Time, error) {
	err := util.CheckNetwork(v.network)
	if err != nil {
		return ContractUpdateStatus{}, nil, err
	}

	v.logger.StartProgress("Checking if contract has been validated")
	defer v.logger.StopProgress()

	return v.getContractValidationStatus(
		contractName,
	)
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

func (v *validator) fetchAndParseReport(reportPath string) ([]ContractUpdateStatus, error) {
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
	var statuses []ContractUpdateStatus
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
