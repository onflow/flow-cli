import Test

access(all) let account = Test.createAccount()

access(all) fun testContract() {
    let err = Test.deployContract(
        name: "Counter",
        path: "../contracts/Counter.cdc",
        arguments: [],
    )

    Test.expect(err, Test.beNil())
}