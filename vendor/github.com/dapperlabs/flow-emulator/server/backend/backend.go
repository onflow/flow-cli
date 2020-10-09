package backend

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/logrusorgru/aurora"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	sdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go/access"
	"github.com/onflow/flow-go/fvm"
	flowgo "github.com/onflow/flow-go/model/flow"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	emulator "github.com/dapperlabs/flow-emulator"
	convert "github.com/dapperlabs/flow-emulator/convert/sdk"
	"github.com/dapperlabs/flow-emulator/types"
)

// Backend wraps an emulated blockchain and implements the RPC handlers
// required by the Access API.
type Backend struct {
	logger   *logrus.Logger
	emulator Emulator
	automine bool
}

// New returns a new backend.
func New(logger *logrus.Logger, emulator Emulator) *Backend {
	return &Backend{
		logger:   logger,
		emulator: emulator,
		automine: false,
	}
}

func (b *Backend) Ping(ctx context.Context) error {
	return nil
}

func (b *Backend) GetNetworkParameters(ctx context.Context) access.NetworkParameters {
	return access.NetworkParameters{
		ChainID: flowgo.Emulator,
	}
}

// GetLatestBlockHeader gets the latest sealed block header.
func (b *Backend) GetLatestBlockHeader(ctx context.Context, isSealed bool) (*flowgo.Header, error) {
	block, err := b.emulator.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	blockID := block.ID()

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Header.Height,
		"blockID":     hex.EncodeToString(blockID[:]),
	}).Debug("🎁  GetLatestBlockHeader called")

	return block.Header, nil
}

// GetBlockHeaderByHeight gets a block header by height.
func (b *Backend) GetBlockHeaderByHeight(
	ctx context.Context,
	height uint64,
) (*flowgo.Header, error) {
	block, err := b.emulator.GetBlockByHeight(height)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	blockID := block.ID()

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Header.Height,
		"blockID":     hex.EncodeToString(blockID[:]),
	}).Debug("🎁  GetBlockHeaderByHeight called")

	return block.Header, nil
}

// GetBlockHeaderByID gets a block header by ID.
func (b *Backend) GetBlockHeaderByID(
	ctx context.Context,
	id sdk.Identifier,
) (*flowgo.Header, error) {
	block, err := b.emulator.GetBlockByID(id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	blockID := block.ID()

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Header.Height,
		"blockID":     hex.EncodeToString(blockID[:]),
	}).Debug("🎁  GetBlockHeaderByID called")

	return block.Header, nil
}

// GetLatestBlock gets the latest sealed block.
func (b *Backend) GetLatestBlock(ctx context.Context, isSealed bool) (*flowgo.Block, error) {
	block, err := b.emulator.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	blockID := block.ID()

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Header.Height,
		"blockID":     hex.EncodeToString(blockID[:]),
	}).Debug("🎁  GetLatestBlock called")

	return block, nil
}

// GetBlockByHeight gets a block by height.
func (b *Backend) GetBlockByHeight(
	ctx context.Context,
	height uint64,
) (*flowgo.Block, error) {
	block, err := b.emulator.GetBlockByHeight(height)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	blockID := block.ID()

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Header.Height,
		"blockID":     hex.EncodeToString(blockID[:]),
	}).Debug("🎁  GetBlockByHeight called")

	return block, nil
}

// GetBlockByHeight gets a block by ID.
func (b *Backend) GetBlockByID(
	ctx context.Context,
	id sdk.Identifier,
) (*flowgo.Block, error) {
	block, err := b.emulator.GetBlockByID(id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	blockID := block.ID()

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Header.Height,
		"blockID":     hex.EncodeToString(blockID[:]),
	}).Debug("🎁  GetBlockByID called")

	return block, nil
}

// GetCollectionByID gets a collection by ID.
func (b *Backend) GetCollectionByID(
	ctx context.Context,
	id sdk.Identifier,
) (*sdk.Collection, error) {
	col, err := b.emulator.GetCollection(id)
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

	return col, nil
}

// SendTransaction submits a transaction to the network.
func (b *Backend) SendTransaction(ctx context.Context, tx sdk.Transaction) error {
	err := b.emulator.AddTransaction(tx)
	if err != nil {
		switch t := err.(type) {
		case *emulator.DuplicateTransactionError:
			return status.Error(codes.InvalidArgument, err.Error())
		case *types.FlowError:
			switch t.FlowError.(type) {
			case *fvm.InvalidSignaturePublicKeyDoesNotExistError,
				*fvm.InvalidSignatureVerificationError,
				*fvm.InvalidProposalKeyPublicKeyDoesNotExistError,
				*fvm.InvalidSignaturePublicKeyRevokedError,
				*fvm.InvalidProposalKeyMissingSignatureError,
				*fvm.MissingPayerError,
				*fvm.MissingSignatureError,
				*fvm.InvalidProposalKeySequenceNumberError:

				return status.Error(codes.InvalidArgument, err.Error())
			default:
				return status.Error(codes.Internal, err.Error())
			}
		default:
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		b.logger.
			WithField("txID", tx.ID().String()).
			Debug("️✉️   Transaction submitted")
	}

	if b.automine {
		b.CommitBlock()
	}

	return nil
}

// GetTransaction gets a transaction by ID.
func (b *Backend) GetTransaction(
	ctx context.Context,
	id sdk.Identifier,
) (*sdk.Transaction, error) {
	tx, err := b.emulator.GetTransaction(id)
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

	return tx, nil
}

// GetTransactionResult gets a transaction by ID.
func (b *Backend) GetTransactionResult(
	ctx context.Context,
	id sdk.Identifier,
) (*sdk.TransactionResult, error) {
	result, err := b.emulator.GetTransactionResult(id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	b.logger.
		WithField("txID", id.String()).
		Debugf("📝  GetTransactionResult called")

	return result, nil
}

// GetAccount returns an account by address at the latest sealed block.
func (b *Backend) GetAccount(
	ctx context.Context,
	address sdk.Address,
) (*sdk.Account, error) {
	b.logger.
		WithField("address", address).
		Debugf("👤  GetAccount called")

	account, err := b.getAccount(address)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccountAtLatestBlock returns an account by address at the latest sealed block.
func (b *Backend) GetAccountAtLatestBlock(
	ctx context.Context,
	address sdk.Address,
) (*sdk.Account, error) {
	b.logger.
		WithField("address", address).
		Debugf("👤  GetAccountAtLatestBlock called")

	account, err := b.getAccount(address)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (b *Backend) getAccount(address sdk.Address) (*sdk.Account, error) {
	account, err := b.emulator.GetAccount(address)
	if err != nil {
		switch err.(type) {
		case emulator.NotFoundError:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return account, nil
}

func (b *Backend) GetAccountAtBlockHeight(
	ctx context.Context,
	address sdk.Address,
	height uint64,
) (*sdk.Account, error) {
	panic("implement me")
}

// ExecuteScriptAtLatestBlock executes a script at a the latest block
func (b *Backend) ExecuteScriptAtLatestBlock(
	ctx context.Context,
	script []byte,
	arguments [][]byte,
) ([]byte, error) {
	b.logger.Debugf("👤  ExecuteScriptAtLatestBlock called")

	block, err := b.emulator.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return b.executeScriptAtBlock(script, arguments, block.Header.Height)
}

// ExecuteScriptAtBlockHeight executes a script at a specific block height
func (b *Backend) ExecuteScriptAtBlockHeight(
	ctx context.Context,
	blockHeight uint64,
	script []byte,
	arguments [][]byte,
) ([]byte, error) {
	b.logger.
		WithField("blockHeight", blockHeight).
		Debugf("👤  ExecuteScriptAtBlockHeight called")

	return b.executeScriptAtBlock(script, arguments, blockHeight)
}

// ExecuteScriptAtBlockID executes a script at a specific block ID
func (b *Backend) ExecuteScriptAtBlockID(
	ctx context.Context,
	blockID sdk.Identifier,
	script []byte,
	arguments [][]byte,
) ([]byte, error) {
	b.logger.
		WithField("blockID", blockID).
		Debugf("👤  ExecuteScriptAtBlockID called")

	block, err := b.emulator.GetBlockByID(blockID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return b.executeScriptAtBlock(script, arguments, block.Header.Height)
}

// GetEventsForHeightRange returns events matching a query.
func (b *Backend) GetEventsForHeightRange(
	ctx context.Context,
	eventType string,
	startHeight, endHeight uint64,
) ([]flowgo.BlockEvents, error) {

	err := validateEventType(eventType)
	if err != nil {
		return nil, err
	}

	latestBlock, err := b.emulator.GetLatestBlock()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// if end height is not set, use latest block height
	// if end height is higher than latest, use latest
	if endHeight == 0 || endHeight > latestBlock.Header.Height {
		endHeight = latestBlock.Header.Height
	}

	// check for invalid queries
	if startHeight > endHeight {
		return nil, status.Error(codes.InvalidArgument, "invalid query: start block must be <= end block")
	}

	results := make([]flowgo.BlockEvents, 0)
	eventCount := 0

	for height := startHeight; height <= endHeight; height++ {
		block, err := b.emulator.GetBlockByHeight(height)
		if err != nil {
			switch err.(type) {
			case emulator.NotFoundError:
				return nil, status.Error(codes.NotFound, err.Error())
			default:
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		events, err := b.emulator.GetEventsByHeight(height, eventType)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		flowEvents, err := convert.SDKEventsToFlow(events)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		result := flowgo.BlockEvents{
			BlockID:        block.ID(),
			BlockHeight:    block.Header.Height,
			BlockTimestamp: block.Header.Timestamp,
			Events:         flowEvents,
		}

		results = append(results, result)
		eventCount += len(events)
	}

	b.logger.WithFields(logrus.Fields{
		"eventType":   eventType,
		"startHeight": startHeight,
		"endHeight":   endHeight,
		"eventCount":  eventCount,
	}).Debugf("🎁  GetEventsForHeightRange called")

	return results, nil
}

// GetEventsForBlockIDs returns events matching a set of block IDs.
func (b *Backend) GetEventsForBlockIDs(
	ctx context.Context,
	eventType string,
	blockIDs []sdk.Identifier,
) ([]flowgo.BlockEvents, error) {

	err := validateEventType(eventType)
	if err != nil {
		return nil, err
	}

	results := make([]flowgo.BlockEvents, 0)
	eventCount := 0

	for _, blockID := range blockIDs {
		block, err := b.emulator.GetBlockByID(blockID)
		if err != nil {
			switch err.(type) {
			case emulator.NotFoundError:
				return nil, status.Error(codes.NotFound, err.Error())
			default:
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		events, err := b.emulator.GetEventsByHeight(block.Header.Height, eventType)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		flowEvents, err := convert.SDKEventsToFlow(events)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		result := flowgo.BlockEvents{
			BlockID:        block.ID(),
			BlockHeight:    block.Header.Height,
			BlockTimestamp: block.Header.Timestamp,
			Events:         flowEvents,
		}

		results = append(results, result)
		eventCount += len(events)
	}

	b.logger.WithFields(logrus.Fields{
		"eventType":  eventType,
		"eventCount": eventCount,
	}).Debugf("🎁  GetEventsForBlockIDs called")

	return results, nil
}

func validateEventType(eventType string) error {
	if len(strings.TrimSpace(eventType)) == 0 {
		return status.Error(codes.InvalidArgument, "invalid query: eventType must not be empty")
	}
	return nil
}

// CommitBlock executes the current pending transactions and commits the results in a new block.
func (b *Backend) CommitBlock() {
	block, results, err := b.emulator.ExecuteAndCommitBlock()
	if err != nil {
		b.logger.WithError(err).Error("Failed to commit block")
		return
	}

	for _, result := range results {
		printTransactionResult(b.logger, result)
	}

	blockID := block.ID()

	b.logger.WithFields(logrus.Fields{
		"blockHeight": block.Header.Height,
		"blockID":     hex.EncodeToString(blockID[:]),
	}).Debugf("📦  Block #%d committed", block.Header.Height)
}

// executeScriptAtBlock is a helper for executing a script at a specific block
func (b *Backend) executeScriptAtBlock(script []byte, arguments [][]byte, blockHeight uint64) ([]byte, error) {
	result, err := b.emulator.ExecuteScriptAtBlock(script, arguments, blockHeight)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	printScriptResult(b.logger, result)

	valueBytes, err := jsoncdc.Encode(result.Value)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return valueBytes, nil
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
