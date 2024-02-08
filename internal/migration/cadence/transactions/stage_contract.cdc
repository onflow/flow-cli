import MigrationContractStaging from {{.MigrationContractStaging}}

/// This transaction is used to stage a contract update for Cadence 1.0 contract migrations.
///
/// Ensure that this transaction is signed by the account that owns the contract to be updated and that the contract
/// has already been deployed to the signing account.
///
/// For more context, see the repo - https://github.com/onflow/contract-updater
///
/// @param contractName: The name of the contract to be updated with the given code
/// @param contractCode: The updated contract code
///
transaction(contractName: String, contractCode: String) {
    let host: &MigrationContractStaging.Host
    
    prepare(signer: AuthAccount) {
        // Configure Host resource if needed
        if signer.borrow<&MigrationContractStaging.Host>(from: MigrationContractStaging.HostStoragePath) == nil {
            signer.save(<-MigrationContractStaging.createHost(), to: MigrationContractStaging.HostStoragePath)
        }
        // Assign Host reference
        self.host = signer.borrow<&MigrationContractStaging.Host>(from: MigrationContractStaging.HostStoragePath)!
    }

    execute {
        // Call staging contract, storing the contract code that will update during Cadence 1.0 migration
        // If code is already staged for the given contract, it will be overwritten.
        MigrationContractStaging.stageContract(host: self.host, name: contractName, code: contractCode)
    }

    post {
        MigrationContractStaging.isStaged(address: self.host.address(), name: contractName):
            "Problem while staging update"
    }
}