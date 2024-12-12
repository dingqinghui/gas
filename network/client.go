/**
 * @Author: dingQingHui
 * @Description:
 * @File: tcpclient
 * @Version: 1.0.0
 * @Date: 2024/12/11 11:19
 */

package network

import (
	"fmt"
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/panjf2000/gnet/v2"
)

func Dial(node api.INode, network, addr string, options ...Option) {
	opts := loadOptions(options...)
	meta := newMeta(node, api.NetConnector, opts)
	switch network {
	case "udp", "udp4", "udp6":
		udpDial(meta, network, addr)
	case "tcp", "tcp4", "tcp6":
		tcpDial(meta, network, addr)
	}
}

func tcpDial(meta *Meta, network, addr string) {
	protoAddr := fmt.Sprintf("%v://%v", network, addr)
	handler := newTcpServer(meta, protoAddr)
	dial(handler, network, addr)
}

func udpDial(meta *Meta, network, addr string) {
	protoAddr := fmt.Sprintf("%v://%v", network, addr)
	server := newUdpServer(meta, protoAddr)
	raw := dial(server, network, addr)
	session := newSession(server, meta, raw)
	server.Link(session, raw)
}

func dial(handler gnet.EventHandler, network, addr string) gnet.Conn {
	client, err := gnet.NewClient(handler)
	xerror.Assert(err)
	xerror.Assert(client.Start())
	raw, err := client.Dial(network, addr)
	xerror.Assert(err)
	return raw
}
