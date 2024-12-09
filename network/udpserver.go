/**
 * @Author: dingQingHui
 * @Description:
 * @File: udpservice
 * @Version: 1.0.0
 * @Date: 2024/11/29 14:48
 */

package network

import (
	"github.com/dingqinghui/gas/api"
	"github.com/panjf2000/gnet/v2"
)

func newUdpServer(node api.INode, serviceType api.NetServerType, protoAddr string, option ...Option) *udpService {
	t := new(udpService)
	t.baseServer = newBaseService(node, serviceType, t, protoAddr, option...)
	return t
}

type udpService struct {
	*baseServer
}

func (t *udpService) Name() string {
	return "udpService:" + t.addr
}

func (t *udpService) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

func (t *udpService) OnTraffic(c gnet.Conn) (action gnet.Action) {
	pid := t.getSessionActor(c)
	if pid == nil {
		_, action = t.baseServer.OnOpen(c)
		if action != gnet.None {
			return
		}
	}
	if pid == nil {
		return gnet.Close
	}
	return t.baseServer.OnTraffic(c)
}

func (t *udpService) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	return gnet.None
}
