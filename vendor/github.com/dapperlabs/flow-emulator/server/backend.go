package server

import (
	"context"
	"fmt"

	"github.com/dapperlabs/flow-go/fvm"
	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jsoncdc "github.com/onflow/cadence/encoding/json"
	sdk "github.com/onflow/flow-go-sdk"
	sdkconvert "github.com/onflow/flow-go-sdk/client/convert"
	"github.com/onflow/flow/protobuf/go/flow/access"
	"github.com/onflow/flow/protobuf/go/flow/entities"

	emulator "github.com/dapperlabs/flow-emulator"
	"github.com/dapperlabs/flow-emulator/types"
)

// Backend wraps an emulated blockchain and implements the RPC handlers
// required by the Observation API.
type Backend struct {
	logger     *logrus.Logger
	blockchain emulator.BlockchainAPI
	automine   bool
}

// NewBackend returns a new backend.
func NewBackend(logger *logrus.Logger, blockchain emulator.BlockchainAPI) *Backend {
	return &Backend{
		logger:     logger,
		blockchain: blockchain,
		automine:   false,
	}
}

// Ping the Observation API server for a response.
func (b *Backend) Ping(ctx context.Context, req *access.PingRequest) (*access.PingResponse, error) {
	return &access.PingResponse{}, nil
}

func (b *Backend) GetNetworkParameters(context.Context, *access.GetNetworkParametersRequest) (*access.GetNetworkParametersResponse, error) {
	panic("implement me")
}

// SendTransaction submits a transaction to the network.
func (b *Backend) SendTransaction(ctx context.Context, req *access.SendTransactionRequest) (*access.SendTransactionResponse, error) {
	txMsg := req.GetTransaction()

	tx, err := sdkconvert.MessageToTransaction(txMsg)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = b.blockchain.AddTransaction(tx)
	if err != nil {
		switch t := err.(type) {
		case *emulator.DuplicateTransactionError:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case *types.FlowError:
			switch t.FlowError.(type) {
			case *fvm.InvalidSignaturePublicKeyError:
				return nil, status.Error(codes.InvalidArgument, err.Error())
			case *fvm.InvalidSignatureAccountError:
				return nil, status.Error(codes.InvalidArgument, err.Error())
			default:
				return nil, status.Error(codes.Internal, err.Error())
			}
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		b.logger.
			WithField("txID", tx.ID().String()).
			Debug("️✉️   Transaction submitted")
	}

	response := &access.SendTransactionResponse{
		Id: tx.ID().Bytes(),
	}

	if b.automine {
		b.commitBlock()
	}

	return response, nil
}

// GetLatestBlockHeader gets the latest sealed block header.
func (b *Backend) GetLatestBlockHeader(ctx context.Context, req *access.GetLatestBlockHeaderRequest) (*access.BlockHeaderResponse, error) {
	block, err := b.blockchain.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID.Hex(),
	}).Debug("🎁  GetLatestBlockHeader called")

	return b.blockToHeaderResponse(block)
}

// GetBlockHeaderByHeight gets a block header by height.
func (b *Backend) GetBlockHeaderByHeight(ctx context.Context, req *access.GetBlockHeaderByHeightRequest) (*access.BlockHeaderResponse, error) {
	block, err := b.blockchain.GetBlockByHeight(req.GetHeight())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID.Hex(),
	}).Debug("🎁  GetBlockHeaderByHeight called")

	return b.blockToHeaderResponse(block)
}

// GetBlockHeaderByID gets a block header by ID.
func (b *Backend) GetBlockHeaderByID(ctx context.Context, req *access.GetBlockHeaderByIDRequest) (*access.BlockHeaderResponse, error) {
	blockID := sdk.HashToID(req.GetId())

	block, err := b.blockchain.GetBlockByID(blockID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID.Hex(),
	}).Debug("🎁  GetBlockHeaderByID called")

	return b.blockToHeaderResponse(block)
}

// GetLatestBlock gets the latest sealed block.
func (b *Backend) GetLatestBlock(ctx context.Context, req *access.GetLatestBlockRequest) (*access.BlockResponse, error) {
	block, err := b.blockchain.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID.Hex(),
	}).Debug("🎁  GetLatestBlock called")

	return b.blockResponse(block)
}

// GetBlockByHeight gets a block by height.
func (b *Backend) GetBlockByHeight(ctx context.Context, req *access.GetBlockByHeightRequest) (*access.BlockResponse, error) {
	block, err := b.blockchain.GetBlockByHeight(req.GetHeight())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID.Hex(),
	}).Debug("🎁  GetBlockByHeight called")

	return b.blockResponse(block)
}

// GetBlockByHeight gets a block by ID.
func (b *Backend) GetBlockByID(ctx context.Context, req *access.GetBlockByIDRequest) (*access.BlockResponse, error) {
	blockID := sdk.HashToID(req.GetId())

	block, err := b.blockchain.GetBlockByID(blockID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID.Hex(),
	}).Debug("🎁  GetBlockByID called")

	return b.blockResponse(block)
}

// GetCollectionByID gets a collection by ID.
func (b *Backend) GetCollectionByID(ctx context.Context, req *access.GetCollectionByIDRequest) (*access.CollectionResponse, error) {
	id := sdk.HashToID(req.GetId())

	col, err := b.blockchain.GetCollection(id)
	if err != nil {
		switch err.(type) {
		case emulator.NotFoundError:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	b.logger.
		WithField("colID", id.Hex()).
		Debugf("📚  GetCollectionByID called")

	return &access.CollectionResponse{
		Collection: sdkconvert.CollectionToMessage(*col),
	}, nil
}

// GetTransaction gets a transaction by ID.
func (b *Backend) GetTransaction(ctx context.Context, req *access.GetTransactionRequest) (*access.TransactionResponse, error) {
	id := sdk.HashToID(req.GetId())

	tx, err := b.blockchain.GetTransaction(id)
	if err != nil {
		switch err.(type) {
		case emulator.NotFoundError:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	b.logger.
		WithField("txID", id.String()).
		Debugf("💵  GetTransaction called")

	txMsg, err := sdkconvert.TransactionToMessage(*tx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &access.TransactionResponse{
		Transaction: txMsg,
	}, nil
}

// GetTransactionResult gets a transaction by ID.
func (b *Backend) GetTransactionResult(ctx context.Context, req *access.GetTransactionRequest) (*access.TransactionResultResponse, error) {
	id := sdk.HashToID(req.GetId())

	result, err := b.blockchain.GetTransactionResult(id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.
		WithField("txID", id.String()).
		Debugf("📝  GetTransactionResult called")

	res, err := sdkconvert.TransactionResultToMessage(*result)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

// GetAccountAtLatestBlock returns an account by address at the latest sealed block.
func (b *Backend) GetAccountAtLatestBlock(
	ctx context.Context,
	req *access.GetAccountAtLatestBlockRequest,
) (*access.AccountResponse, error) {
	address := sdk.BytesToAddress(req.GetAddress())
	account, err := b.blockchain.GetAccount(address)
	if err != nil {
		switch err.(type) {
		case emulator.NotFoundError:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	b.logger.
		WithField("address", address).
		Debugf("👤  GetAccountAtLatestBlock called")

	accMsg := sdkconvert.AccountToMessage(*account)

	return &access.AccountResponse{
		Account: accMsg,
	}, nil
}

func (b *Backend) GetAccountAtBlockHeight(ctx context.Context, request *access.GetAccountAtBlockHeightRequest) (*access.AccountResponse, error) {
	panic("implement me")
}

// ExecuteScriptAtLatestBlock executes a script at a the latest block
func (b *Backend) ExecuteScriptAtLatestBlock(ctx context.Context, req *access.ExecuteScriptAtLatestBlockRequest) (*access.ExecuteScriptResponse, error) {
	script := req.GetScript()
	arguments := req.GetArguments()
	block, err := b.blockchain.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return b.executeScriptAtBlock(script, arguments, block.Height)
}

// ExecuteScriptAtBlockHeight executes a script at a specific block height
func (b *Backend) ExecuteScriptAtBlockHeight(ctx context.Context, req *access.ExecuteScriptAtBlockHeightRequest) (*access.ExecuteScriptResponse, error) {
	script := req.GetScript()
	blockHeight := req.GetBlockHeight()
	arguments := req.GetArguments()
	return b.executeScriptAtBlock(script, arguments, blockHeight)
}

// ExecuteScriptAtBlockID executes a script at a specific block ID
func (b *Backend) ExecuteScriptAtBlockID(ctx context.Context, req *access.ExecuteScriptAtBlockIDRequest) (*access.ExecuteScriptResponse, error) {
	script := req.GetScript()
	arguments := req.GetArguments()
	blockID := sdk.HashToID(req.GetBlockId())

	block, err := b.blockchain.GetBlockByID(blockID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return b.executeScriptAtBlock(script, arguments, block.Height)
}

// GetEventsForHeightRange returns events matching a query.
func (b *Backend) GetEventsForHeightRange(ctx context.Context, req *access.GetEventsForHeightRangeRequest) (*access.EventsResponse, error) {
	startHeight := req.GetStartHeight()
	endHeight := req.GetEndHeight()

	latestBlock, err := b.blockchain.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// if end height is not set, use latest block height
	// if end height is higher than latest, use latest
	if endHeight == 0 || endHeight > latestBlock.Height {
		endHeight = latestBlock.Height
	}

	// check for invalid queries
	if startHeight > endHeight {
		return nil, status.Error(codes.InvalidArgument, "invalid query: start block must be <= end block")
	}

	eventType := req.GetType()

	results := make([]*access.EventsResponse_Result, 0)
	eventCount := 0

	for height := startHeight; height <= endHeight; height++ {
		block, err := b.blockchain.GetBlockByHeight(height)
		if err != nil {
			switch err.(type) {
			case emulator.NotFoundError:
				return nil, status.Error(codes.NotFound, err.Error())
			default:
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		events, err := b.blockchain.GetEventsByHeight(height, eventType)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		result, err := b.eventsBlockResult(block, events)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		results = append(results, result)
		eventCount += len(events)
	}

	b.logger.WithFields(logrus.Fields{
		"eventType":   req.Type,
		"startHeight": req.StartHeight,
		"endHeight":   req.EndHeight,
		"eventCount":  eventCount,
	}).Debugf("🎁  GetEventsForHeightRange called")

	res := access.EventsResponse{
		Results: results,
	}

	return &res, nil
}

// GetEventsForBlockIDs returns events matching a set of block IDs.
func (b *Backend) GetEventsForBlockIDs(ctx context.Context, req *access.GetEventsForBlockIDsRequest) (*access.EventsResponse, error) {
	eventType := req.GetType()

	results := make([]*access.EventsResponse_Result, 0)
	eventCount := 0

	for _, blockID := range req.GetBlockIds() {
		block, err := b.blockchain.GetBlockByID(sdk.HashToID(blockID))
		if err != nil {
			switch err.(type) {
			case emulator.NotFoundError:
				return nil, status.Error(codes.NotFound, err.Error())
			default:
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		events, err := b.blockchain.GetEventsByHeight(block.Height, eventType)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		result, err := b.eventsBlockResult(block, events)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		results = append(results, result)
		eventCount += len(events)
	}

	b.logger.WithFields(logrus.Fields{
		"eventType":  req.Type,
		"eventCount": eventCount,
	}).Debugf("🎁  GetEventsForBlockIDs called")

	res := access.EventsResponse{
		Results: results,
	}

	return &res, nil
}

// commitBlock executes the current pending transactions and commits the results in a new block.
func (b *Backend) commitBlock() {
	block, results, err := b.blockchain.ExecuteAndCommitBlock()
	if err != nil {
		b.logger.WithError(err).Error("Failed to commit block")
		return
	}

	for _, result := range results {
		printTransactionResult(b.logger, result)
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID.Hex(),
	}).Debugf("📦  Block #%d committed", block.Height)
}

// executeScriptAtBlock is a helper for executing a script at a specific block
func (b *Backend) executeScriptAtBlock(script []byte, arguments [][]byte, blockHeight uint64) (*access.ExecuteScriptResponse, error) {
	result, err := b.blockchain.ExecuteScriptAtBlock(script, arguments, blockHeight)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	printScriptResult(b.logger, result)

	valueBytes, err := jsoncdc.Encode(result.Value)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &access.ExecuteScriptResponse{
		Value: valueBytes,
	}

	return response, nil
}

// blockToHeaderResponse constructs a block header response from a block.
func (b *Backend) blockToHeaderResponse(block *sdk.Block) (*access.BlockHeaderResponse, error) {
	msg, err := sdkconvert.BlockHeaderToMessage(*&block.BlockHeader)
	if err != nil {
		return nil, err
	}

	return &access.BlockHeaderResponse{
		Block: msg,
	}, nil
}

// blockResponse constructs a block response from a block.
func (b *Backend) blockResponse(block *sdk.Block) (*access.BlockResponse, error) {
	msg, err := sdkconvert.BlockToMessage(*block)
	if err != nil {
		return nil, err
	}

	return &access.BlockResponse{
		Block: msg,
	}, nil
}

func (b *Backend) eventsBlockResult(
	block *sdk.Block,
	events []sdk.Event,
) (result *access.EventsResponse_Result, err error) {
	eventMessages := make([]*entities.Event, len(events))
	for i, event := range events {
		eventMessages[i], err = sdkconvert.EventToMessage(event)
		if err != nil {
			return nil, err
		}
	}

	blockID := block.ID
	return &access.EventsResponse_Result{
		BlockId:     blockID[:],
		BlockHeight: block.Height,
		Events:      eventMessages,
	}, nil
}

// EnableAutoMine enables the automine flag.
func (b *Backend) EnableAutoMine() {
	b.automine = true
}

// DisableAutoMine disables the automine flag.
func (b *Backend) DisableAutoMine() {
	b.automine = false
}

func printTransactionResult(logger *logrus.Logger, result *types.TransactionResult) {
	if result.Succeeded() {
		logger.
			WithField("txID", result.TransactionID.String()).
			Info("⭐  Transaction executed")
	} else {
		logger.
			WithField("txID", result.TransactionID.String()).
			Warn("❗  Transaction reverted")
	}

	for _, log := range result.Logs {
		logger.Debugf(
			"%s %s",
			logPrefix("LOG", result.TransactionID, aurora.BlueFg),
			log,
		)
	}

	for _, event := range result.Events {
		logger.Debugf(
			"%s %s",
			logPrefix("EVT", result.TransactionID, aurora.GreenFg),
			event,
		)
	}

	if !result.Succeeded() {
		logger.Warnf(
			"%s %s",
			logPrefix("ERR", result.TransactionID, aurora.RedFg),
			result.Error.Error(),
		)
	}
}

func printScriptResult(logger *logrus.Logger, result *types.ScriptResult) {
	if result.Succeeded() {
		logger.
			WithField("scriptID", result.ScriptID.String()).
			Info("⭐  Script executed")
	} else {
		logger.
			WithField("scriptID", result.ScriptID.String()).
			Warn("❗  Script reverted")
	}

	for _, log := range result.Logs {
		logger.Debugf(
			"%s %s",
			logPrefix("LOG", result.ScriptID, aurora.BlueFg),
			log,
		)
	}

	if !result.Succeeded() {
		logger.Warnf(
			"%s %s",
			logPrefix("ERR", result.ScriptID, aurora.RedFg),
			result.Error.Error(),
		)
	}
}

func logPrefix(prefix string, id sdk.Identifier, color aurora.Color) string {
	prefix = aurora.Colorize(prefix, color|aurora.BoldFm).String()
	shortID := fmt.Sprintf("[%s]", id.String()[:6])
	shortID = aurora.Colorize(shortID, aurora.FaintFm).String()
	return fmt.Sprintf("%s %s", prefix, shortID)
}
