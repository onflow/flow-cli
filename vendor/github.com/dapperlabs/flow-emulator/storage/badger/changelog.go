package badger

import (
	"sort"
	"sync"

	"github.com/onflow/flow-go/model/flow"
)

// notFound is a sentinel to indicate that a register has never been written by
// a given block height.
const notFound uint64 = ^uint64(0)

// An ordered list of blocks at which a register changed value. This type
// implements sort.Interface for efficient searching and sorting.
//
// Users should NEVER interact with the backing slice directly, as it must
// be kept sorted for lookups to work.
type changelist struct {
	blocks []uint64
}

// Implement sort.Interface for testing sortedness.
func (c changelist) Len() int { return len(c.blocks) }

func (c changelist) Less(i, j int) bool { return c.blocks[i] < c.blocks[j] }

func (c changelist) Swap(i, j int) { c.blocks[i], c.blocks[j] = c.blocks[j], c.blocks[i] }

// search finds the highest block height B in the changelist so that B<=n.
// Returns notFound if no such block height exists. This relies on the fact
// that the changelist is kept sorted in ascending order.
func (c changelist) search(n uint64) uint64 {
	index := c.searchForIndex(n)
	if index == -1 {
		return notFound
	}
	return c.blocks[index]
}

// searchForIndex finds the index of the highest block height B in the
// changelist so that B<=n. Returns -1 if no such block height exists. This
// relies on the fact that the changelist is kept sorted in ascending order.
func (c changelist) searchForIndex(n uint64) (index int) {
	if len(c.blocks) == 0 {
		return -1
	}
	// This will return the lowest index where the block height is >n.
	// What we want is the index directly BEFORE this.
	foundIndex := sort.Search(c.Len(), func(i int) bool {
		return c.blocks[i] > n
	})

	if foundIndex == 0 {
		// All block heights are >n.
		return -1
	}
	return foundIndex - 1
}

// add adds the block height to the list, ensuring the list remains sorted. If
// n already exists in the list, this is a no-op.
func (c *changelist) add(n uint64) {
	if n == notFound {
		return
	}

	index := c.searchForIndex(n)
	if index == -1 {
		// all blocks in the list are >n, or the list is empty
		c.blocks = append([]uint64{n}, c.blocks...)
		return
	}

	lastBlockHeight := c.blocks[index]
	if lastBlockHeight == n {
		// n already exists in the list
		return
	}

	// insert n directly after lastBlockHeight
	c.blocks = append(c.blocks[:index+1], append([]uint64{n}, c.blocks[index+1:]...)...)
}

// The changelog describes the change history of each register in a ledger.
// For each register, the changelog contains a list of all the block heights at
// which the register's value changed. This enables quick lookups of the latest
// register state change for a given block.
//
// Users of the changelog are responsible for acquiring the mutex before
// reads and writes.
type changelog struct {
	// Maps register IDs to an ordered slice of all the block heights at which
	// the register value changed.
	registers map[string]changelist
	// Guards the register list from concurrent writes.
	sync.RWMutex
}

// newChangelog returns a new changelog.
func newChangelog() changelog {
	return changelog{
		registers: make(map[string]changelist),
		RWMutex:   sync.RWMutex{},
	}
}

// getMostRecentChange returns the most recent block height at which the
// register with the given ID changed value.
func (c changelog) getMostRecentChange(registerID flow.RegisterID, blockHeight uint64) uint64 {
	clist, ok := c.registers[string(registerID)]
	if !ok {
		return notFound
	}

	return clist.search(blockHeight)
}

// changelists returns an exhaustive list of changelists keyed by register ID.
func (c changelog) changelists() map[string]changelist {
	return c.registers
}

// getChangelist returns the changelist corresponding to the given register ID.
// Returns an empty changelist if none exists.
func (c changelog) getChangelist(registerID flow.RegisterID) changelist {
	return c.registers[string(registerID)]
}

// setChangelist sets the changelist for the given register ID, discarding the
// existing changelist if one exists.
func (c changelog) setChangelist(registerID string, clist changelist) {
	c.registers[registerID] = clist
}

// addChange adds a change record to the given register at the given block.
// If the changelist already reflects a change for this register at this block,
// this is a no-op.
//
// If the changelist doesn't exist, it is created.
func (c *changelog) addChange(registerID flow.RegisterID, blockHeight uint64) {
	clist := c.registers[string(registerID)]
	clist.add(blockHeight)
	c.registers[string(registerID)] = clist
}
