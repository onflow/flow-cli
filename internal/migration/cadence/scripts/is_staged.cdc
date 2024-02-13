import "MigrationContractStaging"

/// Returns whether the given contract is staged or not
///
access(all) fun main(contractAddress: Address, contractName: String): Bool {
    return MigrationContractStaging.isStaged(address: contractAddress, name: contractName)
}