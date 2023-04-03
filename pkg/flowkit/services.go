package flowkit

import (
	"context"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
)

//go:generate  mockery --name=Services

type Services interface {
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
	AddContract(context.Context, *Account, *Script, UpdateContract) (flow.Identifier, bool, error)

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
}
