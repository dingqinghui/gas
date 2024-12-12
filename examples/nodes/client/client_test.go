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
	"go.uber.org/zap"
	"testing"
)

type ClientAgent struct {
	//api.AgentActor
	network.AgentActor
}

func (c *ClientAgent) OnInit(ctx api.IActorContext) error {
	_ = c.AgentActor.OnInit(ctx)
	c2s := &common.ClientMessage{
		Name:    "Login",
		Content: "test chat message",
	}
	return c.Session().SendMessage(1, c2s)
}

func (c *ClientAgent) Data(message *common.ClientMessage) error {
	c.Ctx.Info("ClientAgent receive message", zap.Any("message", message))
	return nil
}

func TestNetworkClient(t *testing.T) {
	m := new(common.HandshakeMessage)
	m.Version = "1.1.1"
	handshakeBody := common.MarshalHandshakeMessage(m)

	node := node.New("../../config/client_1.json")
	producer := func() api.IActor { return new(ClientAgent) }

	node.Run()
	router := network.NewRouter()
	router.Register(1, "Data")
	network.Dial(node, "udp", "127.0.0.1:8454",
		network.WithAgentProducer(producer),
		network.WithHandshakeBody(handshakeBody),
		network.WithRouter(router))

	node.Wait()
}
