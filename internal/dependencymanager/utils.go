package dependencymanager

import "github.com/onflow/flow-go/fvm/systemcontracts"

func isCoreContract(contractName string) bool {
	sc := systemcontracts.SystemContractsForChain(flowGo.Emulator)

	for _, coreContract := range sc.All() {
		if coreContract.Name == contractName {
			return true
		}
	}
	return false
}
