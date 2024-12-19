/**
 * @Author: dingQingHui
 * @Description:
 * @File:client
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
	switch network {
	case "udp", "udp4", "udp6":
		udpDial(node, opts, network, addr)
	case "tcp", "tcp4", "tcp6":
		tcpDial(node, opts, network, addr)
	}
}

func tcpDial(node api.INode, opts *Options, network, addr string) {
	protoAddr := fmt.Sprintf("%v://%v", network, addr)
	handler := newTcpServer(node, api.NetConnector, opts, protoAddr)
	dial(handler, network, addr)
}

func udpDial(node api.INode, opts *Options, network, addr string) {
	protoAddr := fmt.Sprintf("%v://%v", network, addr)
	server := newUdpServer(node, api.NetConnector, opts, protoAddr)
	raw := dial(server, network, addr)
	entity := newEntity(server, opts, raw)
	server.Link(entity, raw)
}

func dial(handler gnet.EventHandler, network, addr string) gnet.Conn {
	client, err := gnet.NewClient(handler)
	xerror.Assert(err)
	xerror.Assert(client.Start())
	raw, err := client.Dial(network, addr)
	xerror.Assert(err)
	return raw
}
