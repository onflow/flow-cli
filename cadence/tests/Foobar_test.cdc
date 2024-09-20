import Test

access(all) let account = Test.createAccount()

access(all) fun testContract() {
    let err = Test.deployContract(
        name: "Foobar",
        path: "../contracts/Foobar.cdc",
        arguments: [],
    )

    Test.expect(err, Test.beNil())
}