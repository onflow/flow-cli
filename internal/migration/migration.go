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

package migration

import (
	"bytes"
	"fmt"
	"text/template"
)


const (
	// GetStagedContractCodeFilepath is the file path for the get staged code transaction
	GetStagedContractCodeScriptFilepath = "./cadence/scripts/get_staged_contract_code.cdc"
	// GetStagedCodeForAddressFilepath is the file path for the get all staged contract code for address transaction
	GetStagedCodeForAddressScriptFilepath = "./cadence/scripts/get_all_staged_contract_code_for_address.cdc"
	// IsStagedFilepath is the file path for the is staged transaction
	IsStagedFileScriptpath = "./cadence/scripts/is_staged.cdc"
	// StageContractFilepath is the file path for the stage contract transaction
	StageContractTransactionFilepath = "./cadence/transactions/stage_contract.cdc"
	// UnstageContractFilepath is the file path for the unstage contract transaction
	UnstageContractTransactionFilepath = "./cadence/transactions/unstage_contract.cdc"
)

// TODO: update these  once deployed
var migrationContractStagingAddress = map[string]string{
	"testnet": "0xa983fecbed621163",
	"mainnet": "0xa983fecbed621163",
}

// RenderContractTemplate renders the contract template
func RenderContractTemplate(filepath string, network string) ([]byte, error) {
	scTempl, err := template.ParseFiles(filepath)
	if err != nil {
		return nil, fmt.Errorf("error loading staging contract file: %w", err)
	}

	if migrationContractStagingAddress[network] == "" {
		return nil, fmt.Errorf("staging contract address not found for network: %s", network)
	}	

	// render transaction template with network
	var txScriptBuf bytes.Buffer
	if err := scTempl.Execute(
		&txScriptBuf,
		map[string]string{
			"MigrationContractStaging": migrationContractStagingAddress[network],
		}); err != nil {
		return nil, fmt.Errorf("error rendering staging contract template: %w", err)
	}

	return txScriptBuf.Bytes(), nil 
}