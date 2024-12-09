/**
 * @Author: dingQingHui
 * @Description:
 * @File: api
 * @Version: 1.0.0
 * @Date: 2024/11/12 10:14
 */

package network

import (
	"fmt"
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/netx"
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/panjf2000/gnet/v2"
)

func Dial(node api.INode, network, addr string, opts ...Option) api.INetServer {
	protoAddr := fmt.Sprintf("%v://%v", network, addr)
	r := newService(node, api.NetServiceConnector, protoAddr, opts...)
	client, _ := gnet.NewClient(r)
	xerror.Assert(client.Start())
	_, err := client.Dial(network, addr)
	xerror.Assert(err)
	return r
}

func NewListener(node api.INode, protoAddr string, options ...Option) api.INetServer {
	return newService(node, api.NetServiceListener, protoAddr, options...)
}

func newService(node api.INode, serviceType api.NetServerType, protoAddr string, options ...Option) api.INetServer {
	proto, _, err := netx.ParseProtoAddr(protoAddr)
	xerror.Assert(err)
	switch proto {
	case "udp", "udp4", "udp6":
		return newUdpServer(node, serviceType, protoAddr, options...)
	case "tcp", "tcp4", "tcp6":
		return newTcpServer(node, serviceType, protoAddr, options...)
	}
	return nil
}
