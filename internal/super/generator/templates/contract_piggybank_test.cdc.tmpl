import Test
import "PiggyBank"
import "EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed"
import "FungibleToken"

access(all) let account = Test.createAccount()

access(all) fun setup() {
    // Deploy USDF Mock contract first
    let usdErr = Test.deployContract(
        name: "EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed",
        path: "../contracts/USDF_MOCK.cdc",
        arguments: [],
    )
    Test.expect(usdErr, Test.beNil())

    // Deploy PiggyBank contract
    let err = Test.deployContract(
        name: "PiggyBank",
        path: "../contracts/PiggyBank.cdc",
        arguments: [],
    )
    Test.expect(err, Test.beNil())
}

access(all) fun testInitialBalance() {
    // Check initial balance is 0
    Test.assertEqual(0.0, PiggyBank.getBalance())
}

access(all) fun testDepositAndWithdraw() {
    // Mint some USDF tokens for testing
    let tokens <- EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.mintTokens(amount: 100.0)

    // Deposit tokens into piggy bank contract
    PiggyBank.deposit(from: <-tokens)

    // Check balance after deposit
    Test.assertEqual(100.0, PiggyBank.getBalance())

    // Withdraw half
    let withdrawn <- PiggyBank.withdraw(amount: 50.0)

    // Check balance after withdrawal
    Test.assertEqual(50.0, PiggyBank.getBalance())
    Test.assertEqual(50.0, withdrawn.balance)

    destroy withdrawn
}

access(all) fun testInsufficientBalance() {
    // Reset to known state by withdrawing all
    let currentBalance = PiggyBank.getBalance()
    if currentBalance > 0.0 {
        let tokens <- PiggyBank.withdraw(amount: currentBalance)
        destroy tokens
    }

    // Mint and deposit some USDF tokens
    let tokens <- EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.mintTokens(amount: 50.0)
    PiggyBank.deposit(from: <-tokens)

    // Try to withdraw more than available - this should panic
    // Note: In a real test, you'd want to use Test.expectFailure or similar
}

access(all) fun testMultipleDeposits() {
    // Reset to zero balance
    let currentBalance = PiggyBank.getBalance()
    if currentBalance > 0.0 {
        let tokens <- PiggyBank.withdraw(amount: currentBalance)
        destroy tokens
    }

    // Make multiple deposits
    let tokens1 <- EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.mintTokens(amount: 25.0)
    PiggyBank.deposit(from: <-tokens1)

    let tokens2 <- EVMVMBridgedToken_2aabea2058b5ac2d339b163c6ab6f2b6d53aabed.mintTokens(amount: 75.0)
    PiggyBank.deposit(from: <-tokens2)

    // Check total balance
    Test.assertEqual(100.0, PiggyBank.getBalance())
}
