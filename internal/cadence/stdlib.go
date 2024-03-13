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

package cadence

import (
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/stdlib"
)

type standardLibrary struct {
	baseValueActivation *sema.VariableActivation
}

var _ stdlib.StandardLibraryHandler = standardLibrary{}

func (standardLibrary) ProgramLog(_ string, _ interpreter.LocationRange) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) UnsafeRandom() (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) GetBlockAtHeight(_ uint64) (stdlib.Block, bool, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) GetCurrentBlockHeight() (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) GetAccountBalance(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) GetAccountAvailableBalance(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) CommitStorageTemporarily(_ *interpreter.Interpreter) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) GetStorageUsed(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) GetStorageCapacity(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) GetAccountKey(_ common.Address, _ int) (*stdlib.AccountKey, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) GetAccountContractNames(_ common.Address) ([]string, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) GetAccountContractCode(_ common.AddressLocation) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) EmitEvent(
	_ *interpreter.Interpreter,
	_ *sema.CompositeType,
	_ []interpreter.Value,
	_ interpreter.LocationRange,
) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) AddEncodedAccountKey(_ common.Address, _ []byte) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) RevokeEncodedAccountKey(_ common.Address, _ int) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) AddAccountKey(
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

func (standardLibrary) RevokeAccountKey(_ common.Address, _ int) (*stdlib.AccountKey, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) ParseAndCheckProgram(_ []byte, _ common.Location, _ bool) (*interpreter.Program, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) UpdateAccountContractCode(_ common.AddressLocation, _ []byte) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) RecordContractUpdate(_ common.AddressLocation, _ *interpreter.CompositeValue) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) ContractUpdateRecorded(_ common.AddressLocation) bool {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) InterpretContract(
	_ common.AddressLocation,
	_ *interpreter.Program,
	_ string,
	_ stdlib.DeployedContractConstructorInvocation,
) (*interpreter.CompositeValue, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) TemporarilyRecordCode(_ common.AddressLocation, _ []byte) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) RemoveAccountContractCode(_ common.AddressLocation) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) RecordContractRemoval(_ common.AddressLocation) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) CreateAccount(_ common.Address) (address common.Address, err error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) ValidatePublicKey(_ *stdlib.PublicKey) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) VerifySignature(
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

func (standardLibrary) BLSVerifyPOP(_ *stdlib.PublicKey, _ []byte) (bool, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) Hash(_ []byte, _ string, _ sema.HashAlgorithm) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) AccountKeysCount(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) BLSAggregatePublicKeys(_ []*stdlib.PublicKey) (*stdlib.PublicKey, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) BLSAggregateSignatures(_ [][]byte) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (l standardLibrary) GenerateAccountID(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (l standardLibrary) ReadRandom(_ []byte) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) StartContractAddition(_ common.AddressLocation) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) EndContractAddition(_ common.AddressLocation) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (standardLibrary) IsContractBeingAdded(_ common.AddressLocation) bool {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func newStandardLibrary() (result standardLibrary) {
	result.baseValueActivation = sema.NewVariableActivation(sema.BaseValueActivation)
	for _, valueDeclaration := range stdlib.DefaultStandardLibraryValues(result) {
		result.baseValueActivation.DeclareValue(valueDeclaration)
	}
	return
}

func newScriptStandardLibrary() (result standardLibrary) {
	result.baseValueActivation = sema.NewVariableActivation(sema.BaseValueActivation)
	for _, declaration := range stdlib.DefaultScriptStandardLibraryValues(result) {
		result.baseValueActivation.DeclareValue(declaration)
	}
	return
}
