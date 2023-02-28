package flowkit

import (
	"context"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
)

type Key struct { // todo remove?
	public   crypto.PublicKey
	weight   int
	sigAlgo  crypto.SignatureAlgorithm
	hashAlgo crypto.HashAlgorithm
}

type BlockQuery struct {
	ID     *flow.Identifier
	Height uint64
	Latest bool
}

type EventWorker struct {
	count           int
	blocksPerWorker uint64
}

type Services interface {
	Network() config.Network // I think this would be good to have services in context of network
	Ping() (string, error)

	GetAccount(ctx context.Context, address flow.Address) (*flow.Account, error)
	CreateAccount(ctx context.Context, signer *Account, key Key) (*flow.Account, flow.Identifier, error)
	AddContract(ctx context.Context, account *Account, contract *Script, update bool) (flow.Identifier, bool, error)
	RemoveContract(ctx context.Context, account *Account, name string) (flow.Identifier, error)
	GetBlock(ctx context.Context, query BlockQuery) (*flow.Block, error)
	GetCollection(ctx context.Context, ID flow.Identifier) (*flow.Collection, error)
	GetEvents(ctx context.Context, names []string, startHeight uint64, endHeight uint64, worker *EventWorker) ([]flow.BlockEvents, error)
	GenerateKey(ctx context.Context, inputSeed string, sigAlgo crypto.SignatureAlgorithm) (crypto.PrivateKey, error)
	GenerateMnemonicKey(ctx context.Context, derivationPath string, sigAlgo crypto.SignatureAlgorithm) (crypto.PrivateKey, string, error)
	DeployProject(ctx context.Context, update bool) ([]*project.Contract, error)
	ExecuteScript(ctx context.Context, script *Script) (cadence.Value, error)
	GetTransactionByID(ctx context.Context, ID flow.Identifier, waitSeal bool) (*flow.Transaction, *flow.TransactionResult, error)
	GetTransactionsByBlockID(ctx context.Context, blockID flow.Identifier, waitSeal bool) ([]*flow.Transaction, []*flow.TransactionResult, error)
	BuildTransaction(addresses *transactionAddresses, proposerKeyIndex int, script *Script, gasLimit uint64) (*Transaction, error)
	SignTransactionPayload(signer *Account, payload []byte) (*Transaction, error)
	SendSignedTransaction(tx *Transaction) (*flow.Transaction, *flow.TransactionResult, error)
	SendTransaction(accounts *transactionAccountRoles, script *Script, gasLimit uint64) (*flow.Transaction, *flow.TransactionResult, error)
}
