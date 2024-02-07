import "MigrationContractStaging" from {{.MigrationContractStaging}}

/// Returns the code as it is staged or nil if it not currently staged.
///
access(all) fun main(contractAddress: Address, contractName: String): String? {
    return MigrationContractStaging.getStagedContractCode(address: contractAddress, name: contractName)
}