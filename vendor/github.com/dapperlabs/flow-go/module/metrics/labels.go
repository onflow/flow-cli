package metrics

const (
	LabelChannel  = "topic"
	LabelChain    = "chain"
	EngineLabel   = "engine"
	LabelResource = "resource"
	LabelMessage  = "message"
)

const (
	ChannelOneToOne = "OneToOne"
)

const (
	// collection
	EngineProposal               = "proposal"
	EngineCollectionIngest       = "collection_ingest"
	EngineCollectionProvider     = "collection_provider"
	EngineClusterSynchronization = "cluster-sync"
	// consensus
	EnginePropagation        = "propagation"
	EngineCompliance         = "compliance"
	EngineConsensusProvider  = "consensus_provider"
	EngineConsensusIngestion = "consensus_ingestion"
	EngineMatching           = "matching"
	EngineSynchronization    = "sync"
	// common
	EngineFollower = "follower"
)

const (
	ResourceUndefined            = "undefined"
	ResourceProposal             = "proposal"
	ResourceHeader               = "header"
	ResourceIndex                = "index"
	ResourceIdentity             = "identity"
	ResourceGuarantee            = "guarantee"
	ResourceResult               = "result"
	ResourceReceipt              = "receipt"
	ResourcePendingReceipt       = "pending_receipt" // used at verification node
	ResourceCollection           = "collection"
	ResourcePendingCollection    = "pending_collection" // used at verification node
	ResourceApproval             = "approval"
	ResourceSeal                 = "seal"
	ResourceCommit               = "commit"
	ResourceTransaction          = "transaction"
	ResourceClusterPayload       = "cluster_payload"
	ResourceClusterProposal      = "cluster_proposal"
	ResourceChunkDataPack        = "chunk_data_pack"
	ResourceChunkDataPackTracker = "chunk_data_pack_tracker"
)

const (
	MessageCollectionGuarantee  = "guarantee"
	MessageBlockProposal        = "proposal"
	MessageBlockVote            = "vote"
	MessageExecutionReceipt     = "receipt"
	MessageResultApproval       = "approval"
	MessageSyncRequest          = "ping"
	MessageSyncResponse         = "pong"
	MessageRangeRequest         = "range"
	MessageBatchRequest         = "batch"
	MessageBlockResponse        = "block"
	MessageSyncedBlock          = "synced_block"
	MessageClusterBlockProposal = "cluster_proposal"
	MessageClusterBlockVote     = "cluster_vote"
	MessageClusterBlockResponse = "cluster_block_response"
	MessageSyncedClusterBlock   = "synced_cluster_block"
	MessageTransaction          = "transaction"
	MessageSubmitGuarantee      = "submit_guarantee"
	MessageCollectionRequest    = "collection_request"
	MessageCollectionResponse   = "collection_response"
)
