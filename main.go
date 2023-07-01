package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/chenjia404/p2ptunnel/p2pforwarder"
	"github.com/chenjia404/p2ptunnel/update"
)

var (
	fwr          *p2pforwarder.Forwarder
	fwrCancel    func()
	connections  = make(map[string]func())
	openTCPPorts = make(map[uint]func())
	openUDPPorts = make(map[uint]func())
)

var (
	version   = "0.0.9"
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
	p2p_port := flag.Int("p2p_port", 4001, "p2p use port")
	networkType := flag.String("type", "tcp", "network type tcp/udp")
	var flag_update = flag.Bool("update", false, "update form github")
	flag.Parse()

	var err error

	if *flag_update {
		update.CheckGithubVersion(version)
		return
	}

	fwr, fwrCancel, err = p2pforwarder.NewForwarder(*p2p_port)
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
				log.Println(err)
				//log.Panicln(err)
				//todo 使用中继连接
				//return
			}
		}

		connections[*id] = cancel

		log.Printf("Connections to %s's ports are listened on %s\n", id, listenip)
	}

	select {}
}
