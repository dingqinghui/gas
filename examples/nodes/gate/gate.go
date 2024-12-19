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
	network.AgentActor
}

func (a *ServerAgent) Login(session *api.Session, message *common.ClientMessage) *api.Error {
	a.Ctx.Info("agent receive message", zap.Any("message", message))

	chatPid := a.Ctx.Node().Cluster().NewPid("chat", balancer.NewRandom(), nil)
	if chatPid == nil {
		return api.ErrPidIsNil
	}
	// async inner
	if err := a.Ctx.Send(chatPid, "Join", &common.RpcRoomJoin{UserId: 10}); err != nil {
		return err
	}

	// sync inner
	reply := new(common.RpcRoomJoin)
	if err := a.Ctx.Call(chatPid, "SyncJoin1", &common.RpcRoomJoin{UserId: 11}, reply); err != nil {
		return err
	}

	// network message
	msg := a.Ctx.Message()
	msg.To = chatPid
	msg.MethodName = "Chat"
	if err := a.Ctx.System().PostMessage(chatPid, msg); err != nil {
		return err
	}
	// respond to client
	return a.Ctx.Response(session, message)
}

func HandshakeAuthFunc(entity api.INetEntity, data []byte) ([]byte, *api.Error) {
	m := common.UnmarshalHandshakeMessage(data)
	if m == nil {
		return nil, nil
	}
	entity.Node().Log().Info("HandshakeFuncAuth", zap.String("version", m.Version))
	// 验证客户端信息
	if m.Version == "" {
		if err := entity.Close(errors.New("handshake auth fail")); err != nil {
			entity.Node().Log().Info("HandshakeFuncAuth", zap.Error(err))
			return nil, err
		}
		return nil, api.Ok
	}
	m.ServerTime = time.Now().Unix()
	return common.MarshalHandshakeMessage(m), nil
}

func RunGateNode(path string) {
	gateNode := node.New(path)

	producer := func() api.IActor { return new(ServerAgent) }

	router := network.NewRouters()
	router.Add(1, &network.Router{
		NodeType: "gate",
		ActorId:  0,
		Method:   "Login",
	})
	addrArray := gateNode.GetViper().GetStringSlice("network")
	for _, addr := range addrArray {
		netModule := network.NewListener(gateNode, addr,
			network.WithHandshakeAuth(HandshakeAuthFunc),
			network.WithAgentProducer(producer),
			network.WithRouter(router))

		gateNode.AddModule(netModule)
	}

	gateNode.Run()

	gateNode.Wait()
}
