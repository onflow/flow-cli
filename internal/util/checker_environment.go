/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright Flow Foundation
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

package util

import (
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/ast"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/errors"
	"github.com/onflow/cadence/interpreter"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/sema"
	"github.com/onflow/cadence/stdlib"
	"github.com/onflow/flow-go/fvm/environment"

	"github.com/onflow/flow-go/fvm/evm"
	evmstdlib "github.com/onflow/flow-go/fvm/evm/stdlib"
	"github.com/onflow/flow-go/model/flow"
)

type CheckerEnvironment struct {
	defaultBaseValueActivation     *sema.VariableActivation
	defaultBaseTypeActivation      *sema.VariableActivation
	baseValueActivationsByLocation map[common.Location]*sema.VariableActivation
	baseTypeActivationsByLocation  map[common.Location]*sema.VariableActivation
}

var _ runtime.Environment = CheckerEnvironment{}

func NewCheckerEnvironment() *CheckerEnvironment {
	env := newCheckerEnvironment()
	for _, valueDeclaration := range stdlib.DefaultStandardLibraryValues(env) {
		env.DeclareValue(valueDeclaration, nil)
	}
	return env
}

func NewScriptCheckerEnvironment() *CheckerEnvironment {
	env := newCheckerEnvironment()
	for _, valueDeclaration := range stdlib.DefaultScriptStandardLibraryValues(env) {
		env.DeclareValue(valueDeclaration, nil)
	}
	return env
}

func newCheckerEnvironment() *CheckerEnvironment {
	return &CheckerEnvironment{
		defaultBaseValueActivation:     sema.NewVariableActivation(sema.BaseValueActivation),
		defaultBaseTypeActivation:      sema.NewVariableActivation(sema.BaseTypeActivation),
		baseValueActivationsByLocation: make(map[common.Location]*sema.VariableActivation),
		baseTypeActivationsByLocation:  make(map[common.Location]*sema.VariableActivation),
	}
}

func (e CheckerEnvironment) SetupFVM(chainId flow.ChainID) {
	// Set up the EVM standard library
	evmstdlib.SetupEnvironment(e, nil, evm.ContractAccountAddress(chainId))
}

func (e CheckerEnvironment) GetBaseValueActivation(location common.Location) (baseValueActivation *sema.VariableActivation) {
	if location == nil {
		return e.defaultBaseValueActivation
	}

	baseValueActivation = e.baseValueActivationsByLocation[location]
	if baseValueActivation == nil {
		baseValueActivation = sema.NewVariableActivation(e.defaultBaseValueActivation)
		e.baseValueActivationsByLocation[location] = baseValueActivation
	}
	return
}

func (e CheckerEnvironment) GetBaseTypeActivation(location common.Location) (baseTypeActivation *sema.VariableActivation) {
	if location == nil {
		return e.defaultBaseTypeActivation
	}

	baseTypeActivation = e.baseTypeActivationsByLocation[location]
	if baseTypeActivation == nil {
		baseTypeActivation = sema.NewVariableActivation(e.defaultBaseTypeActivation)
		e.baseTypeActivationsByLocation[location] = baseTypeActivation
	}
	return
}

/*
 * Implement required interface methods
 */

func (e CheckerEnvironment) DeclareValue(valueDeclaration stdlib.StandardLibraryValue, location common.Location) {
	e.GetBaseValueActivation(location).DeclareValue(valueDeclaration)
}

func (e CheckerEnvironment) DeclareType(typeDeclaration stdlib.StandardLibraryType, location common.Location) {
	e.GetBaseTypeActivation(location).DeclareType(typeDeclaration)
}

/*
 * The following methods are not implemented, as they are not used in the type-checking process.
 * They are only used in the execution process.
 */
func (CheckerEnvironment) ProgramLog(_ string, _ interpreter.LocationRange) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) UnsafeRandom() (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) GetBlockAtHeight(_ uint64) (stdlib.Block, bool, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) GetCurrentBlockHeight() (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) GetAccountBalance(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) GetAccountAvailableBalance(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) CommitStorageTemporarily(_ *interpreter.Interpreter) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) GetStorageUsed(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) GetStorageCapacity(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) GetAccountKey(_ common.Address, _ uint32) (*stdlib.AccountKey, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) GetAccountContractNames(_ common.Address) ([]string, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) GetAccountContractCode(_ common.AddressLocation) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) EmitEvent(
	_ *interpreter.Interpreter,
	_ interpreter.LocationRange,
	_ *sema.CompositeType,
	_ []interpreter.Value,
) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) AddEncodedAccountKey(_ common.Address, _ []byte) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) RevokeEncodedAccountKey(_ common.Address, _ int) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) AddAccountKey(
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

func (CheckerEnvironment) RevokeAccountKey(_ common.Address, _ uint32) (*stdlib.AccountKey, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) ParseAndCheckProgram(_ []byte, _ common.Location, _ bool) (*interpreter.Program, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) UpdateAccountContractCode(_ common.AddressLocation, _ []byte) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) RecordContractUpdate(_ common.AddressLocation, _ *interpreter.CompositeValue) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) ContractUpdateRecorded(_ common.AddressLocation) bool {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) InterpretContract(
	_ common.AddressLocation,
	_ *interpreter.Program,
	_ string,
	_ stdlib.DeployedContractConstructorInvocation,
) (*interpreter.CompositeValue, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) TemporarilyRecordCode(_ common.AddressLocation, _ []byte) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) RemoveAccountContractCode(_ common.AddressLocation) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) RecordContractRemoval(_ common.AddressLocation) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) CreateAccount(_ common.Address) (address common.Address, err error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) ValidatePublicKey(_ *stdlib.PublicKey) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) VerifySignature(
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

func (CheckerEnvironment) BLSVerifyPOP(_ *stdlib.PublicKey, _ []byte) (bool, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) Hash(_ []byte, _ string, _ sema.HashAlgorithm) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) AccountKeysCount(_ common.Address) (uint32, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) BLSAggregatePublicKeys(_ []*stdlib.PublicKey) (*stdlib.PublicKey, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) BLSAggregateSignatures(_ [][]byte) ([]byte, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (e CheckerEnvironment) GenerateAccountID(_ common.Address) (uint64, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (e CheckerEnvironment) ReadRandom(_ []byte) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) StartContractAddition(_ common.AddressLocation) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) EndContractAddition(_ common.AddressLocation) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) IsContractBeingAdded(_ common.AddressLocation) bool {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (CheckerEnvironment) SetCompositeValueFunctionsHandler(_ common.TypeID, _ stdlib.CompositeValueFunctionsHandler) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (e CheckerEnvironment) Configure(
	runtimeInterface runtime.Interface,
	codesAndPrograms runtime.CodesAndPrograms,
	storage *runtime.Storage,
	coverageReport *runtime.CoverageReport,
) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (e CheckerEnvironment) Interpret(
	location common.Location,
	program *interpreter.Program,
	f runtime.InterpretFunc,
) (
	interpreter.Value,
	*interpreter.Interpreter,
	error,
) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (e CheckerEnvironment) CommitStorage(inter *interpreter.Interpreter) error {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (e CheckerEnvironment) NewAccountValue(inter *interpreter.Interpreter, address interpreter.AddressValue) interpreter.Value {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (e CheckerEnvironment) DecodeArgument(argument []byte, argumentType cadence.Type) (cadence.Value, error) {
	// Implementation should never be called,
	// only its definition is used for type-checking
	panic(errors.NewUnreachableError())
}

func (e CheckerEnvironment) ResolveLocation(
	identifiers []ast.Identifier,
	location common.Location,
) ([]runtime.ResolvedLocation, error) {
	// TODO:
	var cryptoContractAddress common.Address

	return environment.ResolveLocation(
		identifiers,
		location,
		func(address flow.Address) ([]string, error) {
			// TODO:
			return nil, nil
		},
		cryptoContractAddress,
	)
}
