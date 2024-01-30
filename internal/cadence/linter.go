package cadence

import (
	"fmt"
	"path/filepath"
	"strings"

	cadenceLint "github.com/onflow/cadence-tools/lint"
	cdcTests "github.com/onflow/cadence-tools/test/helpers"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/onflow/cadence/tools/analysis"
	"github.com/onflow/flow-cli/flowkit"
	"golang.org/x/exp/maps"
)

type linter struct {
	checkers map[common.Location]*sema.Checker
	state 	*flowkit.State
	filePaths map[string]bool
}

type convertibleError interface {
	error
	ast.HasPosition
}

func (l *linter) lintFile(
	filePath string,
) (
	[]analysis.Diagnostic,
	error,
) {
	diagnostics := make([]analysis.Diagnostic, 0)
	location := common.StringLocation(filePath)

	code, err := l.state.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	program, err := parser.ParseProgram(nil, code, parser.Config{})
	if err != nil {
		errorDiagnostics, err := maybeProcessConvertableError(err, location)
		if err != nil {
			return nil, err
		}

		diagnostics = append(diagnostics, errorDiagnostics...)
	}

	checker, err := sema.NewChecker(
		program,
		location,
		nil,
		&sema.Config{
			BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
				return sema.NewVariableActivation(sema.BaseValueActivation)
			},
			AccessCheckMode:            sema.AccessCheckModeStrict,
			PositionInfoEnabled:        true,
			ExtendedElaborationEnabled: true,
			LocationHandler:            l.handleLocation,
			ImportHandler: 							l.handleImport,
			AttachmentsEnabled:         true,
		},
	)
	if err != nil {
		return nil, err
	}

	checkError := checker.Check()
	if checkError != nil {
		errorDiagnostics, err := maybeProcessConvertableError(checkError, location)
		if err != nil {
			return nil, err
		}

		diagnostics = append(diagnostics, errorDiagnostics...)
	}

	analysisProgram := analysis.Program{
		Program:     program,
		Checker: 		 checker,
		Location:    checker.Location,
		Code:        []byte(code),
	}

	report := func(diagnostic analysis.Diagnostic) {
		diagnostics = append(diagnostics, diagnostic)
	}

	analyzers := maps.Values(cadenceLint.Analyzers)
	analysisProgram.Run(analyzers, report)

	return diagnostics, nil
}

func maybeProcessConvertableError(
	err error,
	location common.Location,
) (
	[]analysis.Diagnostic,
	error,
) {
	diagnostics := make([]analysis.Diagnostic, 0)
	if parentErr, ok := err.(errors.ParentError); ok {
		checkerDiagnostics, err := getDiagnosticsForParentError(parentErr, location)
		if err != nil {
			return nil, err
		}

		diagnostics = append(diagnostics, checkerDiagnostics...)
	}
	return diagnostics, nil
}

func getDiagnosticsForParentError(err errors.ParentError, location common.Location) ([]analysis.Diagnostic, error) {
	diagnostics := make([]analysis.Diagnostic, 0)

	for _, childErr := range err.ChildErrors() {
		convertibleErr, ok := childErr.(convertibleError)
		if !ok {
			return nil, fmt.Errorf("unable to convert non-convertable error to diagnostic: %T", childErr)
		}
		diagnostic := convertError(convertibleErr, location)
		if diagnostic == nil {
			continue
		}

		diagnostics = append(diagnostics, *diagnostic)
	}

	return diagnostics, nil
}

func convertError(
	err convertibleError,
	location common.Location,
) (
	*analysis.Diagnostic,
) {
	startPosition := err.StartPosition()
	endPosition := err.EndPosition(nil)

	var message string
	var secondaryMessage string

	message = err.Error()
	if secondaryError, ok := err.(errors.SecondaryError); ok {
		secondaryMessage = secondaryError.SecondaryError()
	}

	// TODO: DO THE HASFIXES thing
	suggestedFixes := make([]analysis.SuggestedFix, 0)

	diagnostic := analysis.Diagnostic{
		Location: location,
		Category: "error",
		Message:  message,
		SecondaryMessage: secondaryMessage,
		SuggestedFixes: suggestedFixes,
		Range: ast.Range{
			StartPos: startPosition,
			EndPos:   endPosition,
		},
	}

	/*if _, ok := err.(*sema.ImportedProgramError); ok {
		return nil
	}*/

	return &diagnostic
}

func (l *linter) resolveImport(location common.Location, parentLocation common.Location) (*ast.Program, error) {
	// NOTE: important, *DON'T* return an error when a location type
	// is not supported: the import location can simply not be resolved,
	// no error occurred while resolving it.
	//
	// For example, the Crypto contract has an IdentifierLocation,
	// and we simply return no code for it, so that the checker's
	// import handler is called which resolves the location

	var code []byte
	switch location.(type) {
		case common.StringLocation:
			var filename string
			var err error

			// if the location is not a cadence file try getting the code by identifier
			if !strings.Contains(location.String(), ".cdc") {
				contract, err := l.state.Contracts().ByName(location.String())
				if err != nil {
					return nil, err
				}

				filename = contract.Location
			} else {
				filename = filepath.Join(parentLocation.String(), location.String())
			}

			code, err = l.state.ReadFile(filename)
			if err != nil {
				return nil, err
			}
		default:
			return nil, nil
	}

	return parser.ParseProgram(nil, code, parser.Config{})
}

func (l *linter) handleImport(
	checker *sema.Checker,
	importedLocation common.Location,
	_ ast.Range,
) (
	sema.Import,
	error,
) {
	switch importedLocation {
	case stdlib.CryptoCheckerLocation:
		cryptoChecker := stdlib.CryptoChecker()
		return sema.ElaborationImport{
			Elaboration: cryptoChecker.Elaboration,
		}, nil
	case stdlib.TestContractLocation:
		testChecker := stdlib.GetTestContractType().Checker
		return sema.ElaborationImport{
			Elaboration: testChecker.Elaboration,
		}, nil
	case cdcTests.BlockchainHelpersLocation:
		helpersChecker := cdcTests.BlockchainHelpersChecker()
		return sema.ElaborationImport{
			Elaboration: helpersChecker.Elaboration,
		}, nil
	default:
		importedChecker, ok := l.checkers[importedLocation]
		if !ok {
			importedProgram, err := l.resolveImport(importedLocation, checker.Location)

			if err != nil {
				return nil, err
			}
			if importedProgram == nil {
				return nil, &sema.CheckerError{
					Errors: []error{fmt.Errorf("cannot import %s", importedLocation)},
				}
			}

			importedChecker, err = checker.SubChecker(importedProgram, importedLocation)
			if err != nil {
				return nil, err
			}
			l.checkers[importedLocation] = importedChecker
			err = importedChecker.Check()
			if err != nil {
				return nil, err
			}
		}

		return sema.ElaborationImport{
			Elaboration: importedChecker.Elaboration,
		}, nil
	}			
}

func (l *linter) handleLocation(
	identifiers []ast.Identifier,
	location common.Location,
) (
	[]sema.ResolvedLocation,
	error,
) {
	if _, isAddress := location.(common.AddressLocation); isAddress {
		return nil, fmt.Errorf("address locations are not supported")
	}

	return []runtime.ResolvedLocation{
		{
			Location:    location,
			Identifiers: identifiers,
		},
	}, nil
}