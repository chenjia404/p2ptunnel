package main

import (
	"flag"
	"fmt"
	"log"
	"p2ptunnel/p2pforwarder"
	"runtime"
)

var (
	fwr          *p2pforwarder.Forwarder
	fwrCancel    func()
	connections  = make(map[string]func())
	openTCPPorts = make(map[uint]func())
	openUDPPorts = make(map[uint]func())
)

var (
	version   = "0.0.3"
	gitRev    = ""
	buildTime = ""
)

func main() {

	fmt.Printf("p2ptunnel %s-%s\n", version, gitRev)
	fmt.Printf("buildTime %s\n", buildTime)
	fmt.Printf("System version: %s\n", runtime.GOARCH+"/"+runtime.GOOS)
	fmt.Printf("Golang version: %s\n", runtime.Version())

	port := flag.Uint("l", 12000, "listen port")
	ip := flag.String("ip", "127.0.0.1", "forwarder to ip or listen ip")
	id := flag.String("id", "", "Destination multiaddr id string")
	networkType := flag.String("type", "tcp", "network type tcp/udp")
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
		listenip, cancel, err := fwr.Connect(*id, *ip)
		if err != nil {
			log.Printf("Connect id:%s ip:%s\n", *id, *ip)
			listenip, cancel, err = fwr.Connect(*id, *ip)
			if err != nil {
				log.Panicln(err)
				return
			}
		}

		connections[*id] = cancel

		log.Printf("Connections to %s's ports are listened on %s\n", id, listenip)
	}

	select {}
}
