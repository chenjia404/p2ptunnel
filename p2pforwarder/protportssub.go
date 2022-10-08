package p2pforwarder

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const portssubProtID protocol.ID = "/p2pforwarder/portssub/1.0.0"

const (
	portssubModeManifest  byte = 0x00
	portssubModeSubscribe byte = 0x01
)

type portsManifest struct {
	tcp []uint16
	udp []uint16
}

func setPortsSubHandler(f *Forwarder) {
	f.host.SetStreamHandler(portssubProtID, func(s network.Stream) {
		onInfoFn("'portssub' from " + s.Conn().RemotePeer().Pretty())

		modeBytes := make([]byte, 1)
		_, err := io.ReadFull(s, modeBytes)
		if err != nil {
			s.Reset()
			onErrFn(fmt.Errorf("portssub handler: %s", err))
			return
		}

		switch modeBytes[0] {
		case portssubModeManifest:
			f.portsSubscriptionsMux.Lock()
			subCh := f.portsSubscriptions[s.Conn().RemotePeer()]
			f.portsSubscriptionsMux.Unlock()

			if subCh == nil {
				return
			}

			portsM, err := readPortsManifest(s)
			if err != nil {
				s.Reset()
				onErrFn(err)
				return
			}
			_, err = s.Write([]byte{0x01})
			if err != nil {
				s.Reset()
				onErrFn(err)
				return
			}

			subCh <- portsM

		case portssubModeSubscribe:
			f.portsSubscribersMux.Lock()
			f.portsSubscribers[s.Conn().RemotePeer()] = struct{}{}
			f.portsSubscribersMux.Unlock()

			b := f.createOpenPortsManifestBytes()

			f.sendPortsManifestToSubscriber(s.Conn().RemotePeer(), b)
		}

		s.Close()
	})
}

func (f *Forwarder) publishOpenPortsManifest() {
	b := f.createOpenPortsManifestBytes()

	f.portsSubscribersMux.Lock()
	for peerid := range f.portsSubscribers {
		go f.sendPortsManifestToSubscriber(peerid, b)
	}
	f.portsSubscribersMux.Unlock()
}

func (f *Forwarder) createOpenPortsManifestBytes() []byte {
	f.openPorts.tcp.mux.Lock()
	f.openPorts.udp.mux.Lock()

	lt := len(f.openPorts.tcp.ports)
	lu := len(f.openPorts.udp.ports)

	b := make([]byte, 2+lt*2+2+lu*2)

	var i int

	binary.BigEndian.PutUint16(b[i:i+2], uint16(lt))
	i += 2

	for k := range f.openPorts.tcp.ports {
		binary.BigEndian.PutUint16(b[i:i+2], k)
		i += 2
	}

	binary.BigEndian.PutUint16(b[i:i+2], uint16(lu))
	i += 2

	for k := range f.openPorts.udp.ports {
		binary.BigEndian.PutUint16(b[i:i+2], k)
		i += 2
	}

	f.openPorts.tcp.mux.Unlock()
	f.openPorts.udp.mux.Unlock()

	return b
}

func (f *Forwarder) sendPortsManifestToSubscriber(peerid peer.ID, b []byte) {
	err := f.sendOpenPortsManifestBytes(peerid, b)
	if err == nil {
		return
	}

	onErrFn(err)

	f.portsSubscribersMux.Lock()
	delete(f.portsSubscribers, peerid)
	f.portsSubscribersMux.Unlock()
}

// ErrConnReset = error Connection reset
var ErrConnReset = errors.New("Connection reset")

func (f *Forwarder) sendOpenPortsManifestBytes(peerid peer.ID, b []byte) error {
	s, err := f.host.NewStream(context.Background(), peerid, portssubProtID)
	if err != nil {
		return fmt.Errorf("sendOpenPortsManifestBytes: %s", err)
	}

	_, err = s.Write([]byte{portssubModeManifest})
	if err != nil {
		s.Reset()
		return fmt.Errorf("sendOpenPortsManifestBytes: %s", err)
	}
	_, err = s.Write(b)
	if err != nil {
		s.Reset()
		return fmt.Errorf("sendOpenPortsManifestBytes: %s", err)
	}

	// Test, if connection have been reset or not
	n, err := io.ReadFull(s, make([]byte, 1))
	if err != nil {
		s.Reset()
		return fmt.Errorf("sendOpenPortsManifestBytes: %s", err)
	}

	if n == 0 {
		s.Reset()
		return fmt.Errorf("sendOpenPortsManifestBytes: %s", ErrConnReset)
	}

	s.Close()

	return nil
}

func readPortsManifest(r io.Reader) (portsM *portsManifest, err error) {
	portsM = new(portsManifest)

	portsM.tcp, err = readPortsInManifest(r)
	if err != nil {
		return
	}
	portsM.udp, err = readPortsInManifest(r)
	if err != nil {
		return
	}

	return
}

func readPortsInManifest(r io.Reader) (ports []uint16, err error) {
	portsNumBytes := make([]byte, 2)
	_, err = io.ReadFull(r, portsNumBytes)
	if err != nil {
		return nil, err
	}

	portsNum := int(binary.BigEndian.Uint16(portsNumBytes))

	ports = make([]uint16, portsNum)

	for i := 0; i < portsNum; i++ {
		portBytes := make([]byte, 2)
		_, err = io.ReadFull(r, portBytes)
		if err != nil {
			return nil, fmt.Errorf("readPortsManifest: %s", err)
		}

		ports[i] = binary.BigEndian.Uint16(portBytes)
	}

	return ports, nil
}
