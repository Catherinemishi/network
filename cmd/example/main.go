/*
 *    Copyright 2018 INS Ecosystem
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/insolar/network"
	"github.com/insolar/network/connection"
	"github.com/insolar/network/node"
	"github.com/insolar/network/resolver"
	"github.com/insolar/network/rpc"
	"github.com/insolar/network/store"
	"github.com/insolar/network/transport"

	"github.com/chzyer/readline"
)

func main() {
	var addr = flag.String("addr", "0.0.0.0:0", "IP Address and port to use")
	var bootstrapAddress = flag.String("bootstrap", "", "IP Address and port to bootstrap against")
	var help = flag.Bool("help", false, "Display Help")
	var stun = flag.Bool("stun", true, "Use STUN")

	flag.Parse()

	if *help {
		displayCLIHelp()
		os.Exit(0)
	}

	bootstrapNodes := getBootstrapNodes(bootstrapAddress)

	configuration := network.NewNetworkConfiguration(
		createResolver(*stun),
		connection.NewConnectionFactory(),
		transport.NewUTPTransportFactory(),
		store.NewMemoryStoreFactory(),
		rpc.NewRPCFactory(map[string]rpc.RemoteProcedure{
			"s": send,
		}))
	dhtNetwork, err := configuration.CreateNetwork(*addr, &network.Options{
		BootstrapNodes: bootstrapNodes,
	})
	if err != nil {
		log.Fatalln("Failed to create network:", err.Error())
	}

	defer closeNetwork(configuration)

	ctx := createContext(dhtNetwork)

	go listen(dhtNetwork)
	bootstrap(bootstrapNodes, dhtNetwork)

	handleSignals(configuration)

	repl(dhtNetwork, ctx)
}

func handleSignals(configuration *network.Configuration) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			closeNetwork(configuration)
		}
	}()
}

func createContext(dhtNetwork *network.DHT) network.Context {
	ctx, err := network.NewContextBuilder(dhtNetwork).SetDefaultNode().Build()
	if err != nil {
		log.Fatalln("Failed to create context:", err.Error())
	}
	return ctx
}

func bootstrap(bootstrapNodes []*node.Node, dhtNetwork *network.DHT) {
	if len(bootstrapNodes) > 0 {
		err := dhtNetwork.Bootstrap()
		if err != nil {
			log.Fatalln("Failed to bootstrap network", err.Error())
		}
	}
}

func listen(dhtNetwork *network.DHT) {
	func() {
		err := dhtNetwork.Listen()
		if err != nil {
			log.Fatalln("Listen failed:", err.Error())
		}
	}()
}

func closeNetwork(configuration *network.Configuration) {
	func() {
		err := configuration.CloseNetwork()
		if err != nil {
			log.Fatalln("Failed to close network:", err.Error())
		}
	}()
}

func repl(dhtNetwork *network.DHT, ctx network.Context) {
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer func() {
		errRlClose := rl.Close()
		if errRlClose != nil {
			panic(errRlClose)
		}
	}()
	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF, readline.ErrInterrupt
			break
		}
		input := strings.Split(line, " ")

		switch input[0] {
		case "help":
			displayInteractiveHelp()
		case "findnode":
			doFindNode(input, dhtNetwork, ctx)
		case "info":
			doInfo(dhtNetwork, ctx)
		default:
			doRPC(input, dhtNetwork, ctx)
		}
	}
}

func getBootstrapNodes(bootstrapAddress *string) []*node.Node {
	var bootstrapNodes []*node.Node
	if *bootstrapAddress != "" {
		address, err := node.NewAddress(*bootstrapAddress)
		if err != nil {
			log.Fatalln("Failed to create bootstrap address:", err.Error())
		}
		bootstrapNode := node.NewNode(address)
		bootstrapNodes = append(bootstrapNodes, bootstrapNode)
	}
	return bootstrapNodes
}

func createResolver(stun bool) resolver.PublicAddressResolver {
	var publicAddressResolver resolver.PublicAddressResolver
	if stun {
		publicAddressResolver = resolver.NewStunResolver("")
	} else {
		publicAddressResolver = resolver.NewExactResolver()
	}
	return publicAddressResolver
}

func doFindNode(input []string, dhtNetwork *network.DHT, ctx network.Context) {
	if len(input) != 2 {
		displayInteractiveHelp()
		return
	}
	fmt.Println("Searching for targetNode", input[1])
	targetNode, exists, err := dhtNetwork.FindNode(ctx, input[1])
	if err != nil {
		fmt.Println(err.Error())
	}
	if exists {
		fmt.Println("..Found targetNode:", targetNode)
	} else {
		fmt.Println("..Nothing found for this id!")
	}
}

func doInfo(dhtNetwork *network.DHT, ctx network.Context) {
	nodes := dhtNetwork.NumNodes(ctx)
	originID := dhtNetwork.GetOriginID(ctx)
	fmt.Println("ID: " + originID)
	fmt.Println("Known nodes: " + strconv.Itoa(nodes))
}

func doRPC(input []string, dhtNetwork *network.DHT, ctx network.Context) {
	if len(input) < 2 || len(input[0]) == 0 || len(input[1]) == 0 {
		if len(input) > 0 && len(input[0]) > 0 {
			displayInteractiveHelp()
		}
		return
	}

	method, target := input[0], input[1]
	args := make([][]byte, 0, 4)
	for _, arg := range input[2:] {
		args = append(args, []byte(arg))
	}

	fmt.Printf("Running remote method %s on %s with args %v \n", method, target, args)

	result, err := dhtNetwork.RemoteProcedureCall(ctx, target, method, args)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(string(result))
	}
}

func displayCLIHelp() {
	fmt.Println(`example

Usage:
	example --addr [addr]

Options:
	--help Show this screen.
	--addr=<ip> Local IP and Port [default: 0.0.0.0]
	--bootstrap=<ip> Bootstrap IP and Port
	--stun=<bool> Use STUN protocol for public addr discovery [default: true]`)
}

func displayInteractiveHelp() {
	fmt.Println(`
help - This message
findnode <key> - Find node's real network address
info - Display information about this node

<method> <target> <args...> - Remote procedure call`)
}

func send(sender *node.Node, args [][]byte) ([]byte, error) {
	bs := append([]byte{}, []byte(time.Now().Format(time.Kitchen))...)
	bs = append(bs, ' ')
	bs = append(bs, sender.ID.String()...)

	for _, item := range args {
		bs = append(bs, ' ')
		bs = append(bs, item...)
	}

	fmt.Println(string(bs))

	return bs, nil
}
