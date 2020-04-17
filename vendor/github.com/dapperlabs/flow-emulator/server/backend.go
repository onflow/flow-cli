package server

import (
	"context"
	"fmt"

	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jsoncdc "github.com/dapperlabs/cadence/encoding/json"
	"github.com/dapperlabs/flow-go-sdk"
	"github.com/dapperlabs/flow-go-sdk/client/convert"
	"github.com/dapperlabs/flow/protobuf/go/flow/access"
	"github.com/dapperlabs/flow/protobuf/go/flow/entities"

	emulator "github.com/dapperlabs/flow-emulator"
	emuconvert "github.com/dapperlabs/flow-emulator/server/convert"
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

// SendTransaction submits a transaction to the network.
func (b *Backend) SendTransaction(ctx context.Context, req *access.SendTransactionRequest) (*access.SendTransactionResponse, error) {
	txMsg := req.GetTransaction()

	tx, err := convert.MessageToTransaction(txMsg)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = b.blockchain.AddTransaction(tx)
	if err != nil {
		switch err.(type) {
		case *emulator.DuplicateTransactionError:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case *emulator.InvalidSignaturePublicKeyError:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case *emulator.InvalidSignatureAccountError:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		b.logger.
			WithField("txID", tx.ID().Hex()).
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
		"blockID":     block.ID().Hex(),
	}).Debug("🎁  GetLatestBlockHeader called")

	return b.blockToHeaderResponse(block), nil
}

// GetBlockHeaderByHeight gets a block header by height.
func (b *Backend) GetBlockHeaderByHeight(ctx context.Context, req *access.GetBlockHeaderByHeightRequest) (*access.BlockHeaderResponse, error) {
	block, err := b.blockchain.GetBlockByHeight(req.GetHeight())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID().Hex(),
	}).Debug("🎁  GetBlockHeaderByHeight called")

	return b.blockToHeaderResponse(block), nil
}

// GetBlockHeaderByID gets a block header by ID.
func (b *Backend) GetBlockHeaderByID(ctx context.Context, req *access.GetBlockHeaderByIDRequest) (*access.BlockHeaderResponse, error) {
	blockID := flow.HashToID(req.GetId())

	block, err := b.blockchain.GetBlockByID(blockID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID().Hex(),
	}).Debug("🎁  GetBlockHeaderByID called")

	return b.blockToHeaderResponse(block), nil
}

// GetLatestBlock gets the latest sealed block.
func (b *Backend) GetLatestBlock(ctx context.Context, req *access.GetLatestBlockRequest) (*access.BlockResponse, error) {
	block, err := b.blockchain.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID().Hex(),
	}).Debug("🎁  GetLatestBlock called")

	return b.blockResponse(block), nil
}

// GetBlockByHeight gets a block by height.
func (b *Backend) GetBlockByHeight(ctx context.Context, req *access.GetBlockByHeightRequest) (*access.BlockResponse, error) {
	block, err := b.blockchain.GetBlockByHeight(req.GetHeight())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID().Hex(),
	}).Debug("🎁  GetBlockByHeight called")

	return b.blockResponse(block), nil
}

// GetBlockByHeight gets a block by ID.
func (b *Backend) GetBlockByID(ctx context.Context, req *access.GetBlockByIDRequest) (*access.BlockResponse, error) {
	blockID := flow.HashToID(req.GetId())

	block, err := b.blockchain.GetBlockByID(blockID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Height,
		"blockID":     block.ID().Hex(),
	}).Debug("🎁  GetBlockByID called")

	return b.blockResponse(block), nil
}

// GetCollectionByID gets a collection by ID.
func (b *Backend) GetCollectionByID(ctx context.Context, req *access.GetCollectionByIDRequest) (*access.CollectionResponse, error) {
	id := flow.HashToID(req.GetId())

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
		Collection: emuconvert.CollectionToMessage(*col),
	}, nil
}

// GetTransaction gets a transaction by ID.
func (b *Backend) GetTransaction(ctx context.Context, req *access.GetTransactionRequest) (*access.TransactionResponse, error) {
	id := flow.HashToID(req.GetId())

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
		WithField("txID", id.Hex()).
		Debugf("💵  GetTransaction called")

	return &access.TransactionResponse{
		Transaction: convert.TransactionToMessage(*tx),
	}, nil
}

// GetTransactionResult gets a transaction by ID.
func (b *Backend) GetTransactionResult(ctx context.Context, req *access.GetTransactionRequest) (*access.TransactionResultResponse, error) {
	id := flow.HashToID(req.GetId())

	result, err := b.blockchain.GetTransactionResult(id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.
		WithField("txID", id.Hex()).
		Debugf("📝  GetTransactionResult called")

	res, err := convert.TransactionResultToMessage(*result)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

// GetAccount returns the info associated with an address.
func (b *Backend) GetAccount(ctx context.Context, req *access.GetAccountRequest) (*access.GetAccountResponse, error) {
	address := flow.BytesToAddress(req.GetAddress())
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
		Debugf("👤  GetAccount called")

	accMsg, err := convert.AccountToMessage(*account)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &access.GetAccountResponse{
		Account: accMsg,
	}, nil
}

// ExecuteScriptAtLatestBlock executes a script at a the latest block
func (b *Backend) ExecuteScriptAtLatestBlock(ctx context.Context, req *access.ExecuteScriptAtLatestBlockRequest) (*access.ExecuteScriptResponse, error) {
	script := req.GetScript()
	block, err := b.blockchain.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return b.executeScriptAtBlock(script, block.Height)
}

// ExecuteScriptAtBlockHeight executes a script at a specific block height
func (b *Backend) ExecuteScriptAtBlockHeight(ctx context.Context, req *access.ExecuteScriptAtBlockHeightRequest) (*access.ExecuteScriptResponse, error) {
	script := req.GetScript()
	blockHeight := req.GetBlockHeight()
	return b.executeScriptAtBlock(script, blockHeight)
}

// ExecuteScriptAtBlockID executes a script at a specific block ID
func (b *Backend) ExecuteScriptAtBlockID(ctx context.Context, req *access.ExecuteScriptAtBlockIDRequest) (*access.ExecuteScriptResponse, error) {
	script := req.GetScript()
	blockID := flow.HashToID(req.GetBlockId())

	block, err := b.blockchain.GetBlockByID(blockID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return b.executeScriptAtBlock(script, block.Height)
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
		block, err := b.blockchain.GetBlockByID(flow.HashToID(blockID))
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
		"blockID":     block.ID().Hex(),
	}).Debugf("📦  Block #%d committed", block.Height)
}

// executeScriptAtBlock is a helper for executing a script at a specific block
func (b *Backend) executeScriptAtBlock(script []byte, blockHeight uint64) (*access.ExecuteScriptResponse, error) {
	result, err := b.blockchain.ExecuteScriptAtBlock(script, blockHeight)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	printScriptResult(b.logger, result)

	if result.Value == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid script")
	}

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
func (b *Backend) blockToHeaderResponse(block *types.Block) *access.BlockHeaderResponse {
	return &access.BlockHeaderResponse{
		Block: convert.BlockHeaderToMessage(block.Header()),
	}
}

// blockResponse constructs a block response from a block.
func (b *Backend) blockResponse(block *types.Block) *access.BlockResponse {
	return &access.BlockResponse{
		Block: emuconvert.BlockToMessage(*block),
	}
}

func (b *Backend) eventsBlockResult(
	block *types.Block,
	events []flow.Event,
) (result *access.EventsResponse_Result, err error) {
	eventMessages := make([]*entities.Event, len(events))
	for i, event := range events {
		eventMessages[i], err = convert.EventToMessage(event)
		if err != nil {
			return nil, err
		}
	}

	return &access.EventsResponse_Result{
		BlockId:     block.ID().Bytes(),
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

func printTransactionResult(logger *logrus.Logger, result *emulator.TransactionResult) {
	if result.Succeeded() {
		logger.
			WithField("txID", result.TransactionID.Hex()).
			Info("⭐  Transaction executed")
	} else {
		logger.
			WithField("txID", result.TransactionID.Hex()).
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
			event.String(),
		)
	}

	if result.Reverted() {
		logger.Warnf(
			"%s %s",
			logPrefix("ERR", result.TransactionID, aurora.RedFg),
			result.Error.Error(),
		)
	}
}

func printScriptResult(logger *logrus.Logger, result *emulator.ScriptResult) {
	if result.Succeeded() {
		logger.
			WithField("scriptID", result.ScriptID.Hex()).
			Info("⭐  Script executed")
	} else {
		logger.
			WithField("scriptID", result.ScriptID.Hex()).
			Warn("❗  Script reverted")
	}

	for _, log := range result.Logs {
		logger.Debugf(
			"%s %s",
			logPrefix("LOG", result.ScriptID, aurora.BlueFg),
			log,
		)
	}

	for _, event := range result.Events {
		logger.Debugf(
			"%s %s",
			logPrefix("EVT", result.ScriptID, aurora.GreenFg),
			event.String(),
		)
	}

	if result.Reverted() {
		logger.Warnf(
			"%s %s",
			logPrefix("ERR", result.ScriptID, aurora.RedFg),
			result.Error.Error(),
		)
	}
}

func logPrefix(prefix string, id flow.Identifier, color aurora.Color) string {
	prefix = aurora.Colorize(prefix, color|aurora.BoldFm).String()
	shortID := fmt.Sprintf("[%s]", id.Hex()[:6])
	shortID = aurora.Colorize(shortID, aurora.FaintFm).String()
	return fmt.Sprintf("%s %s", prefix, shortID)
}
