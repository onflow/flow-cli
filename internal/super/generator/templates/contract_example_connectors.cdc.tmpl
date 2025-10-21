// Contract: ExampleConnectors
// Purpose: Minimal example of a DeFiActions Sink implementation that accepts any
//          fungible token vault type and deposits it into a provided vault capability.
//
// Concepts demonstrated:
// - Implementing DeFiActions.Sink with type-safe deposits
// - Using a capability to a fungible vault for deposits
// - Exposing component metadata and optional UniqueIdentifier wiring
//
// Safety:
// - depositCapacity enforces type equality with a precondition
// - Withdrawals are sized by callers via minimumCapacity() or DeFiActions patterns
import "FungibleToken"
import "DeFiActions"

access(all) contract ExampleConnectors {
    // TokenSink: A simple Sink that deposits everything it receives
    // NOTE: In practice, this should not be used as it is essentially a duplicate
    // of the standard FungibleTokenConnectors.VaultSink connector.
    access(all) struct TokenSink: DeFiActions.Sink {
        // Capability to a receiver that can accept deposits of the matching vault type
        access(contract) let vault: Capability<&{FungibleToken.Receiver}>
        // Optional tracing ID used by DeFiActions to correlate flows
        access(contract) var uniqueID: DeFiActions.UniqueIdentifier?

        init(
            vault: Capability<&{FungibleToken.Receiver}>,
            uniqueID: DeFiActions.UniqueIdentifier?
        ) {
            self.vault = vault
            self.uniqueID = uniqueID
        }

        // Required by Sink: advertise the exact deposit type supported
        access(all) view fun getSinkType(): Type {
            return self.vault.borrow()!.getType()
        }

        // This sink places no limit on deposits; callers may size with their own rules
        access(all) fun minimumCapacity(): UFix64 {
            return UFix64.max
        }

        // Deposit the full balance from the provided vault into the target vault
        access(all) fun depositCapacity(from: auth(FungibleToken.Withdraw) &{FungibleToken.Vault}) {
            pre {
                // Enforce exact type match between provided vault and sink type
                from.getType() == self.getSinkType():
                "Invalid vault provided for deposit - \(from.getType().identifier) is not \(self.getSinkType().identifier)"
            }
            // No-op for empty transfers
            let amount: UFix64 = from.balance
            if amount == 0.0 { return }
            // Move all funds and deposit
            let payment <- from.withdraw(amount: amount)
            self.vault.borrow()!.deposit(from: <-payment)
        }

        // Report metadata about this component for DeFiActions graph inspection
        access(all) fun getComponentInfo(): DeFiActions.ComponentInfo {
            return DeFiActions.ComponentInfo(
                type: self.getType(),
                id: self.id(),
                innerComponents: []
            )
        }

        // Implementation detail for UniqueIdentifier passthrough
        access(contract) view fun copyID(): DeFiActions.UniqueIdentifier? {
            return self.uniqueID
        }

        // Allow the framework to set/propagate a UniqueIdentifier for tracing
        access(contract) fun setID(_ id: DeFiActions.UniqueIdentifier?) {
            self.uniqueID = id
        }
    }
}

