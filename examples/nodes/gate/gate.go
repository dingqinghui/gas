/**
 * @Author: dingQingHui
 * @Description:
 * @File: gate_test
 * @Version: 1.0.0
 * @Date: 2024/11/25 10:15
 */

package gate

import (
	"errors"
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/cluster/balancer"
	"github.com/dingqinghui/gas/examples/common"
	"github.com/dingqinghui/gas/network"
	"github.com/dingqinghui/gas/node"
	"go.uber.org/zap"
	"time"
)

type ServerAgent struct {
	api.BuiltinActor
}

func (a *ServerAgent) Data(message *common.ClientMessage, respond *common.ClientMessage) error {
	respond.Content = "1111111111111111"
	a.Ctx.Info("agent receive message", zap.Any("message", message))
	// 外部消息分发
	switch message.Name {
	case "Login":
		return a.Login(message)
	}
	return nil
}

func (a *ServerAgent) Login(request *common.ClientMessage) error {
	// todo 登录验证
	pid := api.NewPidWithName("ChatService")
	msg := &common.RpcRoomJoin{
		UserId: 1, // fake
	}
	// rpc
	if err := a.Ctx.Send(pid, "Join", msg); err != nil {
		return err
	}

	return nil
}

func (a *ServerAgent) JoinSuccess(message *common.ClientMessage) error {
	a.Ctx.Info("JoinSuccess")
	return nil
}

func HandshakeAuthFunc(ctx api.IActorContext, data []byte) ([]byte, error) {
	m := common.UnmarshalHandshakeMessage(data)
	ctx.Info("HandshakeFuncAuth", zap.String("version", m.Version))
	// 验证客户端信息
	if m.Version == "" {
		if err := ctx.Process().AsyncStop(); err != nil {
			ctx.Info("HandshakeFuncAuth", zap.Error(err))
			return nil, err
		}
		return nil, errors.New("handshake auth fail")
	}

	m.ServerTime = time.Now().Unix()
	return common.MarshalHandshakeMessage(m), nil
}

func RunGateNode(path string) {
	gateNode := node.New(path)

	gateNode.Cluster().SetLB("ChatService", balancer.NewRandom())
	producer := func() api.IActor { return new(ServerAgent) }

	router := network.NewRouter()
	router.Register(1, "Data")
	addrArray := gateNode.GetViper().GetStringSlice("network")
	for _, addr := range addrArray {
		netModule := network.NewListener(gateNode, addr,
			network.WithHandshakeAuth(HandshakeAuthFunc),
			network.WithAgentProducer(producer),
			network.WithRouter(router))

		gateNode.AddModule(netModule)
	}

	gateNode.Run()
}
