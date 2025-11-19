import Test
import "FungibleToken"
import "FlowToken"
import "ExampleConnectors"
import "DeFiActions"

access(all) let serviceAccount = Test.serviceAccount()
access(all) let recipient = Test.createAccount()

access(all) fun setup() {
    // Deploy DeFi Actions dependencies first
    // Note: The Cadence Test Framework requires manual deployment of dependencies
    // unless used in Fork Testing mode.
    var err = Test.deployContract(
        name: "DeFiActionsUtils",
        path: "../../imports/6d888f175c158410/DeFiActionsUtils.cdc",
        arguments: []
    )
    Test.expect(err, Test.beNil())

    err = Test.deployContract(
        name: "DeFiActions",
        path: "../../imports/6d888f175c158410/DeFiActions.cdc",
        arguments: []
    )
    Test.expect(err, Test.beNil())

    // Deploy ExampleConnectors
    err = Test.deployContract(
        name: "ExampleConnectors",
        path: "../contracts/ExampleConnectors.cdc",
        arguments: []
    )
    Test.expect(err, Test.beNil())
}

access(all) fun testTokenSinkDeposit() {
    // Execute transaction to test TokenSink
    // Service account already has FLOW tokens
    let code = Test.readFile("../transactions/DepositViaSink.cdc")
    let tx = Test.Transaction(
        code: code,
        authorizers: [serviceAccount.address],
        signers: [serviceAccount],
        arguments: [recipient.address, 10.0]
    )

    let result = Test.executeTransaction(tx)
    Test.expect(result, Test.beSucceeded())
}

