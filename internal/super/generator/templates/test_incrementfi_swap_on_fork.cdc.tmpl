// Fork Testing Example
//
// The #test_fork pragma configures this test to run against a snapshot of mainnet state.
// This allows testing against real deployed contracts and production data without deploying
// anything to a live network. All mutations happen locally in your test environment.
//
// To run this test: flow test cadence/tests/test_incrementfi_swap_on_fork.cdc
//
// Learn more about fork testing in the README.md file or at:
// https://developers.flow.com/blockchain-development-tutorials/cadence/fork-testing
#test_fork(network: "mainnet-fork", height: nil)

import Test
import "DeFiActions"
import "FlowToken"
import "stFlowToken"

// Executes a minimal swap from FLOW -> stFlow using IncrementFi on a forked mainnet.
// Withdraws a tiny amount from a known FLOW holder and swaps via IncrementFi router.
access(all) fun testIncrementFi_SwapOnFork() {
	// Arbitrary mainnet account that has FLOW balance and vaults already setup
	// Fork testing allows impersonating any mainnet account for testing
	let HOLDER = Test.getAccount(0x42a06f24a1049154)
	let AMOUNT_IN: UFix64 = 0.001

	let txCode = Test.readFile("../transactions/incrementfi_swap.cdc")

	// Define vault types for FLOW -> stFlow swap
	let flowVaultType = Type<@FlowToken.Vault>()
	let stFlowVaultType = Type<@stFlowToken.Vault>()

	let res = Test.executeTransaction(
		Test.Transaction(
			code: txCode,
			authorizers: [HOLDER.address],
			signers: [HOLDER],
			arguments: [AMOUNT_IN, flowVaultType, stFlowVaultType]
		)
	)

	Test.expect(res, Test.beSucceeded())

	// Log all swap events emitted during the transaction
	let swapEvents = Test.eventsOfType(Type<DeFiActions.Swapped>())
	log("Swap events:")
	for event in swapEvents {
		log(event)
	}
}


