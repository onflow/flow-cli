import "MigrationContractStaging"

/// Returns the code for all staged contracts hosted by the given contract address.
///
access(all) fun main(contractAddress: Address): {String: String} {
    return MigrationContractStaging.getAllStagedContractCode(forAddress: contractAddress)
}