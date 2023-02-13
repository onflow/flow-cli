package test

import (
	"fmt"
	cdcTests "github.com/onflow/cadence-tools/test"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

func importResolver(
	scriptPath string,
	readerWriter flowkit.ReaderWriter,
	contracts config.Contracts,
) cdcTests.ImportResolver {
	return func(location common.Location) (string, error) {
		stringLocation, isFileImport := location.(common.StringLocation)
		if !isFileImport {
			return "", fmt.Errorf("cannot import from %s", location)
		}

		importedContract, err := resolveContract(contracts, stringLocation)
		if err != nil {
			return "", err
		}

		importedContractFilePath := util.AbsolutePath(scriptPath, importedContract.Location)

		contractCode, err := readerWriter.ReadFile(importedContractFilePath)
		if err != nil {
			return "", err
		}

		return string(contractCode), nil
	}
}

func resolveContract(
	contracts config.Contracts,
	stringLocation common.StringLocation,
) (config.Contract, error) {
	relativePath := stringLocation.String()
	for _, contract := range contracts {
		if contract.Location == relativePath {
			return contract, nil
		}
	}

	return config.Contract{},
		fmt.Errorf("cannot find contract with location '%s' in configuration", relativePath)
}

func fileResolver(
	scriptPath string,
	readerWriter flowkit.ReaderWriter,
) cdcTests.FileResolver {
	return func(path string) (string, error) {
		importFilePath := util.AbsolutePath(scriptPath, path)

		content, err := readerWriter.ReadFile(importFilePath)
		if err != nil {
			return "", err
		}

		return string(content), nil
	}
}
