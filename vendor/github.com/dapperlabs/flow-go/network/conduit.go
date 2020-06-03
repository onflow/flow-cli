// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package network

import (
	"github.com/dapperlabs/flow-go/model/flow"
)

// Conduit represents the interface for engines to communicate over the
// peer-to-peer network. Upon registration with the network, each engine is
// assigned a conduit, which it can use to communicate across the network in
// a network-agnostic way. In the background, the network layer connects all
// engines with the same ID over a shared bus, accessible through the conduit.
type Conduit interface {

	// Submit will submit end event to the network layer. The network layer will
	// ensure that the event is delivered to the same engine on the desired target
	// nodes. It's possible that the event traverses other nodes than the target
	// nodes on its path across the network. The network codec needs to be aware
	// of how to encode the given event type, otherwise the send will fail.
	Submit(event interface{}, targetIDs ...flow.Identifier) error
}
