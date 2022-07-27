package blocks

import (
	"testing"
	"time"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

type MockBlocks struct{}

func (e *MockBlocks) GetBlock(
	query string,
	eventType string,
	verbose bool,
) (*flow.Block, []flow.BlockEvents, []*flow.Collection, error) {
	return &flow.Block{
		BlockHeader: flow.BlockHeader{
			ID:        [32]byte{'h', 'i'},
			ParentID:  [32]byte{'n', 'o'},
			Height:    123,
			Timestamp: time.Now(),
		},
		BlockPayload: flow.BlockPayload{
			CollectionGuarantees: []*flow.CollectionGuarantee{},
			Seals:                []*flow.BlockSeal{},
		},
	}, []flow.BlockEvents{}, []*flow.Collection{}, nil
}

// GetLatestBlockHeight returns the latest block height
func (e *MockBlocks) GetLatestBlockHeight() (uint64, error) {
	// TODO
	return 1, nil
}

func Test_blocks(t *testing.T) {
	t.Run("Get Latest Block command", func(t *testing.T) {

		s := &services.Services{
			Blocks: &MockBlocks{},
		}

		readerWriter := tests.ReaderWriter()
		res, err := get([]string{"latest"}, readerWriter, command.Flags, s)
		if err != nil {
			panic(err)
		}

		expected := "Block ID\t6162000000000000000000000000000000000000000000000000000000000000\nParent ID\t6162000000000000000000000000000000000000000000000000000000000000\nTimestamp\t2022-07-27 14:27:49.585751 -0600 MDT m=+0.025544111\nHeight\t123\nTotal Seals\t0\nTotal Collections\t0"

		assert.Equal(t, res.String(), expected, "Get command output does not match the expected output")
	})
}
