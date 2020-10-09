package backend

import (
	"context"

	"github.com/onflow/flow-go/access"
	flowgo "github.com/onflow/flow-go/model/flow"

	convert "github.com/dapperlabs/flow-emulator/convert/sdk"
)

// Adapter wraps the emulator backend to be compatible with access.API.
type Adapter struct {
	backend *Backend
}

// NewAdapter returns a new backend adapter.
func NewAdapter(backend *Backend) *Adapter {
	return &Adapter{backend: backend}
}

func (a *Adapter) Ping(ctx context.Context) error {
	return a.backend.Ping(ctx)
}

func (a *Adapter) GetNetworkParameters(ctx context.Context) access.NetworkParameters {
	return a.backend.GetNetworkParameters(ctx)
}

func (a *Adapter) GetLatestBlockHeader(ctx context.Context, isSealed bool) (*flowgo.Header, error) {
	return a.backend.GetLatestBlockHeader(ctx, isSealed)
}

func (a *Adapter) GetBlockHeaderByHeight(ctx context.Context, height uint64) (*flowgo.Header, error) {
	return a.backend.GetBlockHeaderByHeight(ctx, height)
}

func (a *Adapter) GetBlockHeaderByID(ctx context.Context, id flowgo.Identifier) (*flowgo.Header, error) {
	return a.backend.GetBlockHeaderByID(ctx, convert.FlowIdentifierToSDK(id))
}

func (a *Adapter) GetLatestBlock(ctx context.Context, isSealed bool) (*flowgo.Block, error) {
	return a.backend.GetLatestBlock(ctx, isSealed)
}

func (a *Adapter) GetBlockByHeight(ctx context.Context, height uint64) (*flowgo.Block, error) {
	return a.backend.GetBlockByHeight(ctx, height)
}

func (a *Adapter) GetBlockByID(ctx context.Context, id flowgo.Identifier) (*flowgo.Block, error) {
	return a.backend.GetBlockByID(ctx, convert.FlowIdentifierToSDK(id))
}

func (a *Adapter) GetCollectionByID(ctx context.Context, id flowgo.Identifier) (*flowgo.LightCollection, error) {
	collection, err := a.backend.GetCollectionByID(ctx, convert.FlowIdentifierToSDK(id))
	if err != nil {
		return nil, err
	}

	return convert.SDKCollectionToFlow(collection), nil
}

func (a *Adapter) SendTransaction(ctx context.Context, tx *flowgo.TransactionBody) error {
	return a.backend.SendTransaction(ctx, convert.FlowTransactionToSDK(*tx))
}

func (a *Adapter) GetTransaction(ctx context.Context, id flowgo.Identifier) (*flowgo.TransactionBody, error) {
	tx, err := a.backend.GetTransaction(ctx, convert.FlowIdentifierToSDK(id))
	if err != nil {
		return nil, err
	}

	return convert.SDKTransactionToFlow(*tx), nil
}

func (a *Adapter) GetTransactionResult(ctx context.Context, id flowgo.Identifier) (*access.TransactionResult, error) {
	result, err := a.backend.GetTransactionResult(ctx, convert.FlowIdentifierToSDK(id))
	if err != nil {
		return nil, err
	}

	flowResult, err := convert.SDKTransactionResultToFlow(result)
	if err != nil {
		return nil, err
	}

	return flowResult, nil
}

func (a *Adapter) GetAccount(ctx context.Context, address flowgo.Address) (*flowgo.Account, error) {
	account, err := a.backend.GetAccount(ctx, convert.FlowAddressToSDK(address))
	if err != nil {
		return nil, err
	}

	flowAccount, err := convert.SDKAccountToFlow(account)
	if err != nil {
		return nil, err
	}

	return flowAccount, nil
}

func (a *Adapter) GetAccountAtLatestBlock(ctx context.Context, address flowgo.Address) (*flowgo.Account, error) {
	account, err := a.backend.GetAccountAtLatestBlock(ctx, convert.FlowAddressToSDK(address))
	if err != nil {
		return nil, err
	}

	flowAccount, err := convert.SDKAccountToFlow(account)
	if err != nil {
		return nil, err
	}

	return flowAccount, nil
}

func (a *Adapter) GetAccountAtBlockHeight(
	ctx context.Context,
	address flowgo.Address,
	height uint64,
) (*flowgo.Account, error) {
	account, err := a.backend.GetAccountAtBlockHeight(ctx, convert.FlowAddressToSDK(address), height)
	if err != nil {
		return nil, err
	}

	flowAccount, err := convert.SDKAccountToFlow(account)
	if err != nil {
		return nil, err
	}

	return flowAccount, nil
}

func (a *Adapter) ExecuteScriptAtLatestBlock(
	ctx context.Context,
	script []byte,
	arguments [][]byte,
) ([]byte, error) {
	return a.backend.ExecuteScriptAtLatestBlock(ctx, script, arguments)
}

func (a *Adapter) ExecuteScriptAtBlockHeight(
	ctx context.Context,
	blockHeight uint64,
	script []byte,
	arguments [][]byte,
) ([]byte, error) {
	return a.backend.ExecuteScriptAtBlockHeight(ctx, blockHeight, script, arguments)
}

func (a *Adapter) ExecuteScriptAtBlockID(
	ctx context.Context,
	blockID flowgo.Identifier,
	script []byte,
	arguments [][]byte,
) ([]byte, error) {
	return a.backend.ExecuteScriptAtBlockID(ctx, convert.FlowIdentifierToSDK(blockID), script, arguments)
}

func (a *Adapter) GetEventsForHeightRange(
	ctx context.Context,
	eventType string,
	startHeight, endHeight uint64,
) ([]flowgo.BlockEvents, error) {
	return a.backend.GetEventsForHeightRange(ctx, eventType, startHeight, endHeight)
}

func (a *Adapter) GetEventsForBlockIDs(
	ctx context.Context,
	eventType string,
	blockIDs []flowgo.Identifier,
) ([]flowgo.BlockEvents, error) {
	return a.backend.GetEventsForBlockIDs(ctx, eventType, convert.FlowIdentifiersToSDK(blockIDs))
}
