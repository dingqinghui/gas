/**
 * @Author: dingQingHui
 * @Description:
 * @File: tcpservice
 * @Version: 1.0.0
 * @Date: 2024/11/29 14:13
 */

package network

import (
	"github.com/dingqinghui/gas/api"
)

func newTcpServer(node api.INode, serviceType api.NetServerType, protoAddr string, option ...Option) *tcpServer {
	t := new(tcpServer)
	t.baseServer = newBaseService(node, serviceType, t, protoAddr, option...)
	return t
}

type tcpServer struct {
	*baseServer
}

func (t *tcpServer) Name() string {
	return "tcpServer:" + t.addr
}
