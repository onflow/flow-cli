/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2022 Dapper Labs, Inc.
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

// NOTE: This file is a copy of the file https://github.com/onflow/cadence-tools/blob/master/languageserver/server/stdlib.go

package util

import (
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/stdlib"
)

type StandardLibrary struct {
	BaseValueActivation *sema.VariableActivation
}

var _ stdlib.StandardLibraryHandler = StandardLibrary{}

func (StandardLibrary) ProgramLog(_ string, _ interpreter.LocationRange) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) UnsafeRandom() (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) GetBlockAtHeight(_ uint64) (stdlib.Block, bool, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) GetCurrentBlockHeight() (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) GetAccountBalance(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) GetAccountAvailableBalance(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) CommitStorageTemporarily(_ *interpreter.Interpreter) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) GetStorageUsed(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) GetStorageCapacity(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) GetAccountKey(_ common.Address, _ int) (*stdlib.AccountKey, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) GetAccountContractNames(_ common.Address) ([]string, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) GetAccountContractCode(_ common.AddressLocation) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) EmitEvent(
	_ *interpreter.Interpreter,
	_ *sema.CompositeType,
	_ []interpreter.Value,
	_ interpreter.LocationRange,
) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) AddEncodedAccountKey(_ common.Address, _ []byte) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) RevokeEncodedAccountKey(_ common.Address, _ int) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) AddAccountKey(
	_ common.Address,
	_ *stdlib.PublicKey,
	_ sema.HashAlgorithm,
	_ int,
) (
	*stdlib.AccountKey,
	error,
) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) RevokeAccountKey(_ common.Address, _ int) (*stdlib.AccountKey, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) ParseAndCheckProgram(_ []byte, _ common.Location, _ bool) (*interpreter.Program, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) UpdateAccountContractCode(_ common.AddressLocation, _ []byte) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) RecordContractUpdate(_ common.AddressLocation, _ *interpreter.CompositeValue) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) ContractUpdateRecorded(_ common.AddressLocation) bool {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) InterpretContract(
	_ common.AddressLocation,
	_ *interpreter.Program,
	_ string,
	_ stdlib.DeployedContractConstructorInvocation,
) (*interpreter.CompositeValue, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) TemporarilyRecordCode(_ common.AddressLocation, _ []byte) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) RemoveAccountContractCode(_ common.AddressLocation) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) RecordContractRemoval(_ common.AddressLocation) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) CreateAccount(_ common.Address) (address common.Address, err error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) ValidatePublicKey(_ *stdlib.PublicKey) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) VerifySignature(
	_ []byte,
	_ string,
	_ []byte,
	_ []byte,
	_ sema.SignatureAlgorithm,
	_ sema.HashAlgorithm,
) (bool, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) BLSVerifyPOP(_ *stdlib.PublicKey, _ []byte) (bool, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) Hash(_ []byte, _ string, _ sema.HashAlgorithm) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) AccountKeysCount(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) BLSAggregatePublicKeys(_ []*stdlib.PublicKey) (*stdlib.PublicKey, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) BLSAggregateSignatures(_ [][]byte) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (l StandardLibrary) GenerateAccountID(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (l StandardLibrary) ReadRandom(_ []byte) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) StartContractAddition(_ common.AddressLocation) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) EndContractAddition(_ common.AddressLocation) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (StandardLibrary) IsContractBeingAdded(_ common.AddressLocation) bool {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func NewStandardLibrary() (result StandardLibrary) {
	result.BaseValueActivation = sema.NewVariableActivation(sema.BaseValueActivation)
	for _, valueDeclaration := range stdlib.DefaultStandardLibraryValues(result) {
		result.BaseValueActivation.DeclareValue(valueDeclaration)
	}
	return
}

func NewScriptStandardLibrary() (result StandardLibrary) {
	result.BaseValueActivation = sema.NewVariableActivation(sema.BaseValueActivation)
	for _, declaration := range stdlib.DefaultScriptStandardLibraryValues(result) {
		result.BaseValueActivation.DeclareValue(declaration)
	}
	return
}
