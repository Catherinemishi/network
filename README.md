Insolar – Network
===============
Abstract networking layer

[![Build Status](https://travis-ci.org/insolar/network.svg?branch=master)](https://travis-ci.org/insolar/network)
[![Go Report Card](https://goreportcard.com/badge/github.com/insolar/network)](https://goreportcard.com/report/github.com/insolar/network)
[![GoDoc](https://godoc.org/github.com/insolar/network?status.svg)](https://godoc.org/github.com/insolar/network)

_This project is still in early development state.
It is not recommended to use it in production environment._

Overview
--------

**Insolar** is a blockchain platform developed by INS.

We took [Kademlia DHT](https://en.wikipedia.org/wiki/Kademlia) original specifications and made significant improvements to make it ready
for real world application by enterprises.

#### Key features of our blockchain network layer:
 - *Support of heterogeneous network topology* with different types of nodes being able to communicate with each other.
   In classic peer-to-peer networks, any node can communicate directly with any other node on the network.
   In a real enterprise environment, this condition is often unacceptable for a variety of reasons including security.
 - *Network routing with a node or node group becoming relays* for others nodes.
   The network can continue to function despite various network restrictions such as firewalls, NATs, etc.
 - *Ability to limit number of gateways to corporate node group via relays* to keep the node group secure while being
   able to interact with the rest of the network through relays. This feature mitigates the risk of DDoS attacks.

Key components
--------------

#### **1.** [Transport](https://godoc.org/github.com/insolar/network/transport)
Network transport interface. It allows to abstract our network from physical transport.
It can either be IP based network or any other kind of message courier (e.g. an industrial message bus). 

#### **2.** [Node](https://godoc.org/github.com/insolar/network/node)
Node is a fundamental part of networking system. Each node has:
 - one real network address (IP or any other transport protocol address)
 - multiple abstract network IDs (either node's own or ones belonging to relayed nodes)

#### **3.** [Routing](https://godoc.org/github.com/insolar/network/routing)
It is actually a Kademlia hash table used to store network nodes and calculate distances between them.
See [Kademlia whitepaper](https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf) and
[XLattice design specification](http://xlattice.sourceforge.net/components/protocol/kademlia/specs.html) for details.

#### **4.** [Message](https://godoc.org/github.com/insolar/network/message)
A set of data transferred by this module between nodes.
 - Request message
 - Response message
 
 Now messages are serialized simply with encoding/gob.
 In future there will be a powerful robust serialization system based on Google's Protocol Buffers.

#### **5.** [RPC](https://godoc.org/github.com/insolar/network/rpc)
RPC module allows higher level components to register methods that can be called by other network nodes.

Installation
------------

To `install` the package, use the command:

    go get github.com/insolar/network

Usage
-----

To get started, use the following `package code`:

```go
package main

import (
	"github.com/insolar/network"
	"github.com/insolar/network/connection"
	"github.com/insolar/network/node"
	"github.com/insolar/network/resolver"
	"github.com/insolar/network/rpc"
	"github.com/insolar/network/store"
	"github.com/insolar/network/transport"
)

func main() {
	configuration := network.NewNetworkConfiguration(
		resolver.NewStunResolver(""),
		connection.NewConnectionFactory(),
		transport.NewUTPTransportFactory(),
		store.NewMemoryStoreFactory(),
		rpc.NewRPCFactory(map[string]rpc.RemoteProcedure{}))

	dhtNetwork, err := configuration.CreateNetwork("0.0.0.0:31337", &network.Options{})
	if err != nil {
		panic(err)
	}
	defer configuration.CloseNetwork()

	dhtNetwork.Listen()
}
```

For more detailed usage example see [cmd/example/main.go](cmd/example/main.go)

Contributing
------------

Please feel free to submit issues, fork the repository and send pull requests!

When submitting an issue, we ask that you please include a complete test function that demonstrates the issue.

**Contact email:** github@ins.world  
**Website:** https://ins.world
