# Flowkit Changelog

All notable changes to flowkit APIs will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Changed

Flowkit package was tagged as v1.0.0 and is now considered stable. 
It was also moved and renamed from `github.com/onflow/pkg/flowkit` to `github.com/onflow/flowkit` (dropped `pkg`).
Please note that when you update to this version you will need to update your imports and 
use the command: `go get github.com/onflow/flowkit@latest` and you will have to manually change all 
imports from `github.com/onflow/pkg/flowkit` to `github.com/onflow/flowkit` (a simple find and replace might do).


--- 

Flowkit package APIs  completely undergo major changes. APIs that were previously accessed via services were moved
under a single interface defined in [services.go](services.go). Accessing those methods must be done on a
flowkit instance.

Example previously:
```go
services := services.NewServices(gateway, state, logger)
account, err := services.Accounts.Get(address)
```
changed to:
```go
services := flowkit.NewFlowkit(state, *network, clientGateway, logger)
account, err := services.GetAccount(context.Background(), address)
```

Each of the APIs now require context as the first argument.
We can find a definition of the flowkit interface here:
```go
Network() config.Network
Ping() error
Gateway() gateway.Gateway
SetLogger(output.Logger)

// GetAccount fetches account on the Flow network.
GetAccount(context.Context, flow.Address) (*flow.Account, error)

// CreateAccount on the Flow network with the provided keys and using the signer for creation transaction.
// Returns the newly created account as well as the ID of the transaction that created the account.
//
// Keys is a slice but only one can be passed as well. If the transaction fails or there are other issues an error is returned.
CreateAccount(context.Context, *Account, []AccountPublicKey) (*flow.Account, flow.Identifier, error)

// AddContract to the Flow account provided and return the transaction ID.
//
// If the contract already exists on the account the operation will fail and error will be returned.
// Use UpdateContract method for such usage.
AddContract(context.Context, *Account, *Script, bool) (flow.Identifier, bool, error)

// RemoveContract from the provided account by its name.
//
// If removal is successful transaction ID is returned.
RemoveContract(context.Context, *Account, string) (flow.Identifier, error)

// GetBlock by the query from Flow blockchain. Query can define a block by ID, block by height or require the latest block.
GetBlock(context.Context, BlockQuery) (*flow.Block, error)

// GetCollection by the ID from Flow network.
GetCollection(context.Context, flow.Identifier) (*flow.Collection, error)

// GetEvents from Flow network by their event name in the specified height interval defined by start and end inclusive.
// Optional worker defines parameters for how many concurrent workers do we want to fetch our events,
// and how many blocks between the provided interval each worker fetches.
//
// Providing worker value will produce faster response as the interval will be scanned concurrently. This parameter is optional,
// if not provided only a single worker will be used.
GetEvents(context.Context, []string, uint64, uint64, *EventWorker) ([]flow.BlockEvents, error)

// GenerateKey using the signature algorithm and optional seed. If seed is not provided a random safe seed will be generated.
GenerateKey(context.Context, crypto.SignatureAlgorithm, string) (crypto.PrivateKey, error)

// GenerateMnemonicKey will generate a new key with the signature algorithm and optional derivation path.
//
// If the derivation path is not provided a default "m/44'/539'/0'/0/0" will be used.
GenerateMnemonicKey(context.Context, crypto.SignatureAlgorithm, string) (crypto.PrivateKey, string, error)

DerivePrivateKeyFromMnemonic(context.Context, string, crypto.SignatureAlgorithm, string) (crypto.PrivateKey, error)

// DeployProject contracts to the Flow network or update if already exists and update is set to true.
//
// Retrieve all the contracts for specified network, sort them for deployment deploy one by one and replace
// the imports in the contract source, so it corresponds to the account name the contract was deployed to.
DeployProject(context.Context, bool) ([]*project.Contract, error)

// ExecuteScript on the Flow network and return the Cadence value as a result.
ExecuteScript(context.Context, *Script) (cadence.Value, error)

// GetTransactionByID from the Flow network including the transaction result. Using the waitSeal we can wait for the transaction to be sealed.
GetTransactionByID(context.Context, flow.Identifier, bool) (*flow.Transaction, *flow.TransactionResult, error)

GetTransactionsByBlockID(context.Context, flow.Identifier) ([]*flow.Transaction, []*flow.TransactionResult, error)

// BuildTransaction builds a new transaction type for later signing and submitting to the network.
//
// TransactionAddressesRoles type defines the address for each role (payer, proposer, authorizers) and the script defines the transaction content.
BuildTransaction(context.Context, *TransactionAddressesRoles, int, *Script, uint64) (*Transaction, error)

// SignTransactionPayload will use the signer account provided and the payload raw byte content to sign it.
//
// The payload should be RLP encoded transaction payload and is suggested to be used in pair with BuildTransaction function.
SignTransactionPayload(context.Context, *Account, []byte) (*Transaction, error)

// SendSignedTransaction will send a prebuilt and signed transaction to the Flow network.
//
// You can build the transaction using the BuildTransaction method and then sign it using the SignTranscation method.
SendSignedTransaction(context.Context, *Transaction) (*flow.Transaction, *flow.TransactionResult, error)

// SendTransaction will build and send a transaction to the Flow network, using the accounts provided for each role and
// contain the script. Transaction as well as transaction result will be returned in case the transaction is successfully submitted.
SendTransaction(context.Context, *TransactionAccountRoles, *Script, uint64) (*flow.Transaction, *flow.TransactionResult, error)
```

Passing network parameter is no longer required as the network is initialized using the `NewFlowkit` initializer.

---

The `ScriptQuery` was moved to `flowkit` package and addded `Latest` field. It is also now passed by value instead of 
by pointer. You can use `LatestScriptQuery` convenience variable to pass usage of latest block for script execution.

----

The `TransactionAccountRoles` and `TransactionAddressesRoles` are no longer passed to transaction functions as 
pointer but as value since they are always required.

---

The `flowkit.NewScript` method was removed and you should use `flowkit.Script{}` struct directly for initialization. 
Also getter and setters were removed to favour direct property access.

---

The `config.Contracts.ByName(name string) *Contract` changed to return an error if contract 
was not found whereas before it returned a nil value.

---

The `Account` type and helper methods as well as `AccountKey` interface and all the implementations
were moved to `accounts` package.

---

The `AccountKey`, `HexAccountKey`, `KmsAccountKey`, `Bip44AccountKey` were renamed to `Key`, `HexKey`, `KMSKey`, `BIP44Key` to avoid stutter in the naming,
as well as all the factory methods were renamed to remove the `account` word.

--- 

The `Transaction` was moved to `transactions` package as well as all the factory methods.
All the factory methods were renamed to drop the `transaction` word.

---

The `TransactionAddressesRoles` was moved to `transactions` package and renamed to drop the 
`transaction` word to `AddressRoles` the same holds true for `TransactionAccountRoles`.

---

The `ParseArgumentsJSON` and `ParseArgumentsWithoutType` were moved to a `arguments` package and renamed 
to `ParseJSON` and `ParseWithoutType` correspondingly.

--- 

Renamed transaction function from `transaction.SetGasLimit` to `transaction.SetComputeLimit`. 


### Added

----
An `AccountPublicKey` type was added used in flowkit `CreateAccount` API, you can find the definition in [flowkit.go](flowkit.go).

---

A `BlockQuery` was added for querying blocks using the flowkit `GetBlock` API. You can find the definition in [flowkit.go](flowkit.go).
You can use `NewBlockQuery` factory method to pass in raw string, which should be either equal to `"latest"`, height or block ID.

### Removed

---

The `GetBlock` API doesn't return the `[]flow.BlockEvents, []*flow.Collection` anymore, you should use `GetCollection` and
`GetEvents` API.

---

The `GetLatestBlockHeight` method was removed, you should instead use `GetBlock(LatestBlockQuery)`.

---

The `flowkit.Exist(path string)` was removed, the `config.Exist(path string)` should be used.

## v0.46.2

### Changed

---
The `Network` property was removed from `Contract` type. The network is now included in
the `Aliases` on the contract. We also removed having multiple contracts by same name just to
accommodate multiple aliases. Now there's only one contract identified by name,
and if there are multiple network aliases they are contained in the `Aliases` list.
- Package: `config`
- Type: `Contracts`

---
A method `Contracts.AddOrUpdate(name, contract)` was changed to not include the name, as it's
already part of the contract you are adding.
- Method: `AddOrUpdate`
- Package: `config`
- Type: `Contracts`

---
Don't return error if contract by name not found but rather just a `nil`.
- Method: `ByName`
- Package: `config`
- Type: `Contracts`

### Added

---

New type `Aliases` was added to `Contracts`.
Aliases contain new functions to get the aliases by network and add new aliases.
- Package: `config`
- Type: `Contracts`


---

`WithLogger` now takes zerolog instead of Logrus since that is what flow-emulator has changed to.
- Package: `gateway`
- Type: `EmulatorGateway`
