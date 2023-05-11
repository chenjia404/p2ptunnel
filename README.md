[English](./README.en.md)

# p2ptunnel

想和朋友联机玩游戏，下班了需要连接公司电脑，但是自己没有服务器，没有公网ip怎么办

本应用可以建立tcp、udp隧道，把本地或者远程应用端口映射出来，不要求有公网，如果双方节点无法进行直连，会有其它节点进行中继转发，数据端对端加密，中继节点无法查看数据。

底层传输可以使用quic、tcp、package websocket、webtransport实现，使用 noise 协议加密传输，自带nat，可以多层组合使用。

## 工作原理

电脑a打开本应用，把一个端口映射出来，电脑b打开本应用，连接电脑a，电脑a的端口映射到本机的127.0.89.1端口下。

如果两台电脑都是内网，会通过其它节点进行数据中继，数据转发的时候会进行端对端加密。

## 使用案例

先下载对于平台的压缩包，解压，然后打开本机的远程桌面。

### 参数说明

|  字段  | 类型 | 说明  |
|  ----  | ----  |----  |
|l  |  ip端口 |转发的本地端口|
|id  | multiaddr格式的 | 连接远程服务id|
|p2p_port|ip端口  |p2p使用的端口，也是监听其它节点连接的端口，默认4001，会自动进行nat，但是可能需要您进行端口映射|
|type|网络类型|tcp或者udp|

### id格式(multiaddr)
|  类型 | 样例|说明  |
|  ----  | ----  |----  |
|12D3KooWLHjy7D    | 纯id| 只知道id，不知道协议、ip这些 |
|/p2p/12D3KooWLHjy7D|纯id | 只知道id，不知道协议、ip这些|
|/ip4/1.1.1.1/tcp/4001/p2p/12D3KooWLHjy7D| 详细路径|知道ip、协议，使用的tcp |
|/ip4/1.1.1.1/udp/4001/quic-v1/p2p/12D3KooWLHjy7D| 详细路径|知道ip、协议，使用的quic |

节点启动的时候会输出相应的地址，把里面的 ip 修改成公网ip即可。

可以通过路径里面的tcp、quic控制连接行为。

### 打开本地端口
`./p2ptunnel -type tcp -l 3389`

注意这里会输出你的节点id，然后通过聊天软件发给你的朋友，这里假设id是12D3。

### 连接
`./p2ptunnel -id 12D3`

连接可能需要几秒到1分钟，连接成功后，会输出 Listening tcp 127.0.89.0:3389 -> 3389

然后朋友在远程桌面连接 127.0.89.0:3389 即可。


### 打包

`goreleaser release --skip-publish  --rm-dist`

## 注意事项

1.本应用虽然使用的端对端加密，但是不保证传输数据的安全性，重要数据请勿使用本应用传递。

2.由于是p2p隧道，所以本程序会连接多个ip，如果介意，请使用frp。

## 上游项目

[go-libp2p](https://github.com/libp2p/go-libp2p)

[p2p-forwarder](https://github.com/nickname32/p2p-forwarder)