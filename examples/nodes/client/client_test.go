/**
 * @Author: dingQingHui
 * @Description:
 * @File: init
 * @Version: 1.0.0
 * @Date: 2024/11/25 10:14
 */

package client

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/examples/common"
	"github.com/dingqinghui/gas/network"
	"github.com/dingqinghui/gas/node"
	"testing"
	"time"
)

type ClientAgent struct {
	//api.AgentActor
	network.AgentActor
}

func (c *ClientAgent) OnInit(ctx api.IActorContext) error {
	time.Sleep(time.Second * 10)
	_ = c.AgentActor.OnInit(ctx)
	c2s := &common.ClientMessage{
		Name:    "Login",
		Content: "test chat message",
	}

	return c.Push(1, c2s)
}

func TestNetworkClient(t *testing.T) {
	m := new(common.HandshakeMessage)
	m.Version = "1.1.1"
	handshakeBody := common.MarshalHandshakeMessage(m)

	node2 := node.New("../../config/client_1.json")
	producer := func() api.IActor { return new(ClientAgent) }
	netComponent := network.Dial(node2, "tcp", "127.0.0.1:8454",
		network.WithAgentProducer(producer), network.WithHandshakeBody(handshakeBody))
	node2.AddModule(netComponent)

	node2.Run()
}
