# p2ptunnel
I want to play games with my friends online, and I need to connect to the company computer after get off work, but I don’t have a server or public IP

This application can establish tcp and udp tunnels to map local or remote application ports. It does not require a public network. If the two nodes cannot be directly connected, there will be other nodes for relay forwarding, data end-to-end encryption, and relay node Unable to view data.

The underlying transmission can be implemented using quic, tcp, package websocket, and webtransport, using the noise protocol to encrypt the transmission, with its own nat, which can be used in multi-layer combinations.

## working principle

Computer a opens the application and maps a port, computer b opens the application, connects to computer a, and the port of computer a is mapped to the port 127.0.89.1 of this machine.

If both computers are intranets, data will be relayed through other nodes, and end-to-end encryption will be performed when data is forwarded.

## Use Cases

First download the compressed package for the platform, unzip it, and then open the remote desktop of the machine.
### Open local port
./p2ptunnel -type tcp -l 3389

Note that your node id will be output here, and then sent to your friends through the chat software, assuming the id is 12D3.

### connection
./p2ptunnel -id 12D3

The connection may take several seconds to 1 minute. After the connection is successful, it will output Listening tcp 127.0.89.0:3389 -> 3389

Then the friend can connect to 127.0.89.0:3389 on the remote desktop.


## note

1. Although this application uses end-to-end encryption, the security of the transmitted data is not guaranteed. Please do not use this application to transfer important data.

2. Because it is a p2p tunnel, this program will connect multiple ips, if you mind, please use frp.

## Upstream project

[go-libp2p](https://github.com/libp2p/go-libp2p)

[p2p-forwarder](https://github.com/nickname32/p2p-forwarder)