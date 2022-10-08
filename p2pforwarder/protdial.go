package p2pforwarder

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"sync"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/pion/udp"
)

const dialProtID protocol.ID = "/p2pforwarder/dial/1.0.0"

var dialsIP = "127.0.88.89"

func setDialHandler(f *Forwarder) {
	f.host.SetStreamHandler(dialProtID, func(s network.Stream) {
		onInfoFn("'dial' from " + s.Conn().RemotePeer().Pretty())

		portBytes := make([]byte, 3)
		_, err := io.ReadFull(s, portBytes)
		if err != nil {
			s.Reset()
			onErrFn(fmt.Errorf("dial handler: %s", err))
			return
		}

		protocolType := portBytes[0]
		port := binary.BigEndian.Uint16(portBytes[1:])

		portInt := int(port)

		var (
			addr string

			portsMap *openPortsStoreMap
		)
		switch protocolType {
		case protocolTypeTCP:
			addr = "tcp:" + strconv.Itoa(portInt)

			portsMap = f.openPorts.tcp
		case protocolTypeUDP:
			addr = "udp:" + strconv.Itoa(portInt)

			portsMap = f.openPorts.udp
		default:
			s.Reset()
			return
		}

		onInfoFn("Dialing to " + addr + " from " + s.Conn().RemotePeer().Pretty())
		defer onInfoFn("Closed dial to " + addr + " from " + s.Conn().RemotePeer().Pretty())

		portsMap.mux.Lock()
		portContext := portsMap.ports[port]
		portsMap.mux.Unlock()

		if portContext == nil {
			s.Reset()
			return
		}

		var conn net.Conn

		switch protocolType {
		case protocolTypeTCP:
			conn, err = net.DialTCP("tcp", &net.TCPAddr{
				IP:   net.ParseIP(dialsIP),
				Port: 0,
			}, &net.TCPAddr{
				IP:   nil,
				Port: portInt,
			})
		case protocolTypeUDP:
			conn, err = net.DialUDP("udp", &net.UDPAddr{
				IP:   net.ParseIP(dialsIP),
				Port: 0,
			}, &net.UDPAddr{
				IP:   nil,
				Port: portInt,
			})
		}

		if err != nil {
			s.Reset()
			onErrFn(fmt.Errorf("dial handler: %s", err))
			return
		}

		pipeBothIOsAndClose(portContext, s, conn)
	})
}

func createAddrInfoString(network string, listenip string, lport int, port int) string {
	return network + " " + listenip + ":" + strconv.Itoa(lport) + " -> " + strconv.Itoa(port)
}

func (f *Forwarder) dial(ctx context.Context, peerid peer.ID, protocolType byte, listenip string, port uint16) {
	lport := int(port)

	var addressinfostr string

	var listenfunc func(lip net.IP, port int) (net.Listener, error)

	switch protocolType {
	case protocolTypeTCP:
		addressinfostr = createAddrInfoString("tcp", listenip, lport, int(port))

		listenfunc = func(lip net.IP, port int) (net.Listener, error) {
			return net.ListenTCP("tcp", &net.TCPAddr{
				IP:   lip,
				Port: port,
			})
		}
	case protocolTypeUDP:
		addressinfostr = createAddrInfoString("udp", listenip, lport, int(port))

		listenfunc = func(lip net.IP, port int) (net.Listener, error) {
			return udp.Listen("udp", &net.UDPAddr{
				IP:   lip,
				Port: port,
			})
		}
	}

	lip := net.ParseIP(listenip)

	ln, err := listenfunc(lip, lport)
	if err != nil {
		onErrFn(fmt.Errorf("dial: %s", err))

		for i := 0; i < 4; i++ {
			lport = rand.Intn(65535-1024) + 1024

			ln, err = listenfunc(lip, lport)

			if err != nil {
				onErrFn(fmt.Errorf("dial: %s", err))
			} else {
				break
			}
		}

		if err != nil {
			return
		}
	}

	onInfoFn("Listening " + addressinfostr)

	go func() {
	loop:
		for {
			conn, err := ln.Accept()
			if err != nil {
				onErrFn(fmt.Errorf("dial: %s", err))
				select {
				case <-ctx.Done():
					break loop
				default:
					continue loop
				}
			}

			onInfoFn("Accepted " + ln.Addr().Network() + " connection from " + conn.RemoteAddr().String() + " on " + ln.Addr().String())

			go func() {
				defer onInfoFn("Closed " + ln.Addr().Network() + " connection from " + conn.RemoteAddr().String() + " on " + ln.Addr().String())

				s, err := f.host.NewStream(ctx, peerid, dialProtID)
				if err != nil {
					conn.Close()
					onErrFn(fmt.Errorf("dial: %s", err))
					return
				}

				p := make([]byte, 3)
				p[0] = protocolType
				binary.BigEndian.PutUint16(p[1:3], port)

				_, err = s.Write(p)
				if err != nil {
					s.Reset()
					conn.Close()
					onErrFn(fmt.Errorf("dial: %s", err))
					return
				}

				pipeBothIOsAndClose(ctx, conn, s)
			}()
		}
	}()

	<-ctx.Done()
	ln.Close()

	onInfoFn("Closed " + addressinfostr)
}

// pipeBothIOsAndClose pipes `a` and `b` in both directions and closes them in the end
func pipeBothIOsAndClose(parentctx context.Context, a io.ReadWriteCloser, b io.ReadWriteCloser) {
	ctx, cancel := context.WithCancel(parentctx)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		wg.Wait()
		cancel()
	}()

	go func() {
		_, err := io.Copy(b, a)
		wg.Done()
		if err != nil {
			onErrFn(fmt.Errorf("pipeBothIOsAndClose b<-a: %s", err))
			cancel()
		}
	}()
	go func() {
		_, err := io.Copy(a, b)
		wg.Done()
		if err != nil {
			onErrFn(fmt.Errorf("pipeBothIOsAndClose a<-b: %s", err))
			cancel()
		}
	}()

	<-ctx.Done()

	a.Close()
	b.Close()
}
