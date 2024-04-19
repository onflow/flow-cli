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
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-github/github"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/migrate/mocks"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/tests"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_IsValidated(t *testing.T) {
	srv, state, _ := util.TestMocks(t)
	testContract := tests.ContractSimple

	// use emulator-account because it already exists in mock, so we don't need to create it
	emuAccount, err := state.EmulatorServiceAccount()
	require.NoError(t, err)

	// Helper function to test the isValidated function
	// with all of the necessary mocks
	testIsValidatedWithStatuses := func(statuses []contractUpdateStatus) (command.Result, error) {
		mockClient := mocks.NewGitHubRepositoriesService(t)

		// mock network
		srv.Network.Return(config.TestnetNetwork)

		// mock github file response
		data, _ := json.Marshal(statuses)
		content := string(data)
		mockFileContent := &github.RepositoryContent{
			Content: &content,
		}
		mockClient.On("GetContents", mock.Anything, "onflow", "cadence", "migrations_data/2.json", mock.Anything).Return(mockFileContent, nil, nil, nil).Once()

		// mock github folder response
		fileType := "file"
		olderPath := "migrations_data/1.json"
		latestPath := "migrations_data/2.json"
		mockFolderContent := []*github.RepositoryContent{
			{
				Path: &olderPath,
				Type: &fileType,
			},
			{
				Path: &latestPath,
				Type: &fileType,
			},
		}
		mockClient.On("GetContents", mock.Anything, "onflow", "cadence", "migrations_data", mock.Anything).Return(nil, mockFolderContent, nil, nil).Once()

		// mock flowkit contract
		state.Contracts().AddOrUpdate(
			config.Contract{
				Name:     testContract.Name,
				Location: testContract.Filename,
			},
		)

		// Add deployment to state
		state.Deployments().AddOrUpdate(
			config.Deployment{
				Network: "testnet",
				Account: emuAccount.Name,
				Contracts: []config.ContractDeployment{
					{
						Name: testContract.Name,
					},
				},
			},
		)

		// call the isValidated function
		res, err := isValidated(mockClient)(
			[]string{testContract.Name},
			command.GlobalFlags{
				Network: "testnet",
			},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.Equal(t, true, mockClient.AssertExpectations(t))
		return res, err
	}

	t.Run("isValidated gets status from latest report on github", func(t *testing.T) {
		res, err := testIsValidatedWithStatuses([]contractUpdateStatus{
			{
				AccountAddress: "0x01",
				ContractName:   "some-other-contract",
				Error:          "4567",
			},
			{
				AccountAddress: emuAccount.Address.HexWithPrefix(),
				ContractName:   testContract.Name,
				Error:          "1234",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)

		expectedUnixTime := 2
		expectedTime := time.Unix(int64(expectedUnixTime), 0)
		require.Equal(t, res.JSON(), validationResult{
			Timestamp: expectedTime,
			Status: contractUpdateStatus{
				AccountAddress: emuAccount.Address.HexWithPrefix(),
				ContractName:   testContract.Name,
				Error:          "1234",
			},
		})
	})

	t.Run("isValidated errors if contract was not in last migration", func(t *testing.T) {
		res, err := testIsValidatedWithStatuses([]contractUpdateStatus{
			{
				AccountAddress: "0x01",
				ContractName:   "some-other-contract",
				Error:          "4567",
			},
		})

		require.ErrorContains(t, err, "has not been part of any emulated migrations yet")
		require.Nil(t, res)
	})
}