// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package engine

import (
	"fmt"
)

// Enum of channel IDs to avoid accidental conflicts.
const (

	// Channels used for testing
	TestNetwork = 00
	TestMetrics = 01

	// Channels for consensus protocols
	ConsensusCommittee = 10
	ConsensusCluster   = 11

	// Channels for protocols actively synchronizing state across nodes
	SyncCommittee = 20
	SyncCluster   = 21
	SyncExecution = 22

	// Channels for actively pushing entities to subscribers
	PushTransactions = 100
	PushGuarantees   = 101
	PushBlocks       = 102
	PushReceipts     = 103
	PushApprovals    = 104

	// Channels for actively requesting missing entities
	RequestCollections       = 200
	RequestChunks            = 201
	RequestReceiptsByBlockID = 202

	// Channel aliases to make the code more readable / more robust to errors
	ReceiveTransactions = PushTransactions
	ReceiveGuarantees   = PushGuarantees
	ReceiveBlocks       = PushBlocks
	ReceiveReceipts     = PushReceipts
	ReceiveApprovals    = PushApprovals

	ProvideCollections       = RequestCollections
	ProvideChunks            = RequestChunks
	ProvideReceiptsByBlockID = RequestReceiptsByBlockID
)

func ChannelName(channelID uint8) string {
	switch channelID {
	case TestNetwork:
		return "test-network"
	case TestMetrics:
		return "test-metrics"
	case ConsensusCommittee:
		return "consensus-committee"
	case ConsensusCluster:
		return "consensus-cluster"
	case SyncCommittee:
		return "sync-committee"
	case SyncCluster:
		return "sync-cluster"
	case SyncExecution:
		return "sync-execution"
	case PushTransactions:
		return "push-transactions"
	case PushGuarantees:
		return "push-guarantees"
	case PushBlocks:
		return "push-blocks"
	case PushReceipts:
		return "push-receipts"
	case PushApprovals:
		return "push-approvals"
	case RequestCollections:
		return "request-collections"
	case RequestChunks:
		return "request-chunks"
	case RequestReceiptsByBlockID:
		return "request-receipts-by-block-id"
	}
	return fmt.Sprintf("unknown-channel-%d", channelID)
}

// FullyQualifiedChannelName returns the unique channel name made up of channel name string suffixed with root block id
// The root block id is used to prevent cross talks between nodes on different sporks
func FullyQualifiedChannelName(channelID uint8, rootBlockID string) string {
	return fmt.Sprintf("%s/%s", ChannelName(channelID), rootBlockID)
}
