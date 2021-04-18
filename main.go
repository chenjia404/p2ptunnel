package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"

	p2pforwarder "github.com/nickname32/p2p-forwarder"
)

var (
	fwr          *p2pforwarder.Forwarder
	fwrCancel    func()
	connections  = make(map[string]func())
	openTCPPorts = make(map[uint]func())
	openUDPPorts = make(map[uint]func())
)

var (
	version   = "0.0.1"
	gitRev    = ""
	buildTime = ""
)

func main() {

	fmt.Printf("p2ptunnel %s-%s\n", version, gitRev)
	fmt.Printf("buildTime %s\n", buildTime)
	fmt.Printf("System version: %s\n", runtime.GOARCH+"/"+runtime.GOOS)
	fmt.Printf("Golang version: %s\n", runtime.Version())

	port := flag.Uint("l", 12000, "libp2p listen port")
	id := flag.String("id", "", "Destination multiaddr id string")
	networkType := flag.String("type", "", "network type tcp/udp")
	flag.Parse()

	var err error

	fwr, fwrCancel, err = p2pforwarder.NewForwarder()
	if err != nil {
		log.Panicln(err)
	}

	if *id == "" {

		cancel, err := fwr.OpenPort(*networkType, uint16(*port))
		if err != nil {
			log.Panicln(err)
			return
		}

		log.Println("Your id: " + fwr.ID())

		switch *networkType {
		case "tcp":
			openTCPPorts[*port] = cancel
		case "udp":
			openUDPPorts[*port] = cancel
		}
	} else {
		listenip, cancel, err := fwr.Connect(*id)
		if err != nil {
			log.Panicln(err)
			return
		}

		connections[*id] = cancel

		log.Println("Connections to %s's ports are listened on %s", id, listenip)
	}

	select {}
}
