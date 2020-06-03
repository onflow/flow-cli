package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/module/trace"
)

// Verification spans.
const chunkExecutionSpanner = "chunk_execution_duration"

type VerificationCollector struct {
	tracer                  *trace.OpenTracer
	chunksCheckedPerBlock   prometheus.Counter
	resultApprovalsPerBlock prometheus.Counter
	storagePerChunk         prometheus.Gauge
}

func NewVerificationCollector(tracer *trace.OpenTracer, registerer prometheus.Registerer, log zerolog.Logger) *VerificationCollector {

	chunksCheckedPerBlock := promauto.NewCounter(prometheus.CounterOpts{
		Name:      "checked_chunks_total",
		Namespace: namespaceVerification,
		Help:      "total number of chunks checked",
	})
	resultApprovalsPerBlock := promauto.NewCounter(prometheus.CounterOpts{
		Name:      "result_approvals_total",
		Namespace: namespaceVerification,
		Help:      "total number of emitted result approvals",
	})
	storagePerChunk := promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "storage_latest_chunk_size_bytes",
		Namespace: namespaceVerification,
		Help:      "latest ingested chunk resources storage (bytes)",
	})

	err := registerer.Register(chunksCheckedPerBlock)
	if err != nil {
		log.Debug().Err(err).Msg("could not register chunksCheckedPerBlock metric")
	}
	err = registerer.Register(resultApprovalsPerBlock)
	if err != nil {
		log.Debug().Err(err).Msg("could not register resultApprovalsPerBlock metric")
	}
	err = registerer.Register(storagePerChunk)
	if err != nil {
		log.Debug().Err(err).Msg("could not register storagePerChunk metric")
	}

	vc := &VerificationCollector{
		tracer:                  tracer,
		chunksCheckedPerBlock:   chunksCheckedPerBlock,
		resultApprovalsPerBlock: resultApprovalsPerBlock,
		storagePerChunk:         storagePerChunk,
	}

	return vc
}

// OnResultApproval is called whenever a result approval is emitted.
// It increases the result approval counter for this chunk.
func (vc *VerificationCollector) OnResultApproval() {
	// increases the counter of disseminated result approvals
	// fo by one. Each result approval corresponds to a single chunk of the block
	// the approvals disseminated by verifier engine
	vc.resultApprovalsPerBlock.Inc()

}

// OnVerifiableChunkSubmitted is called whenever a verifiable chunk is shaped for a specific
// chunk. It adds the size of the verifiable chunk to the histogram. A verifiable chunk is assumed
// to capture all the resources needed to verify a chunk.
// The purpose of this function is to track the overall chunk resources size on disk.
// Todo wire this up to do monitoring
// https://github.com/dapperlabs/flow-go/issues/3183
func (vc *VerificationCollector) OnVerifiableChunkSubmitted(size float64) {
	vc.storagePerChunk.Set(size)
}

// OnChunkVerificationStarted is called whenever the verification of a chunk is started.
// It starts the timer to record the execution time.
func (vc *VerificationCollector) OnChunkVerificationStarted(chunkID flow.Identifier) {
	// starts spanner tracer for this chunk ID
	vc.tracer.StartSpan(chunkID, chunkExecutionSpanner)
}

// OnChunkVerificationFinished is called whenever chunkID verification gets finished.
// It finishes recording the duration of execution and increases number of checked chunks.
func (vc *VerificationCollector) OnChunkVerificationFinished(chunkID flow.Identifier) {
	vc.tracer.FinishSpan(chunkID, chunkExecutionSpanner)
	// increases the checked chunks counter
	// checked chunks are the ones with a chunk data pack disseminated from
	// ingest to verifier engine
	vc.chunksCheckedPerBlock.Inc()

}
