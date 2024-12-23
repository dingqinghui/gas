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
	"github.com/dingqinghui/gas/zlog"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"time"
)

type ServerAgent struct {
	network.AgentActor
}

func (a *ServerAgent) Login(session *api.Session, message *common.ClientMessage) *api.Error {
	zlog.Info("agent receive message", zap.Any("message", message))

	cluster := a.Ctx.System().Node().Cluster()

	chatPid := cluster.NewPid("chat", balancer.NewRandom(), nil)
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
	if err := a.Ctx.Response(session, message); err != nil {
		return nil
	}

	return a.Ctx.Push(session, 1, message)
}

func HandshakeAuthFunc(entity api.INetEntity, data []byte) ([]byte, *api.Error) {
	m := common.UnmarshalHandshakeMessage(data)
	if m == nil {
		return nil, nil
	}
	zlog.Info("HandshakeFuncAuth", zap.String("version", m.Version))
	// 验证客户端信息
	if m.Version == "" {
		if err := entity.Close(errors.New("handshake auth fail")); err != nil {
			zlog.Info("HandshakeFuncAuth", zap.Error(err))
			return nil, err
		}
		return nil, api.Ok
	}
	m.ServerTime = time.Now().Unix()
	return common.MarshalHandshakeMessage(m), nil
}

var routers = network.NewRouters()

func NetRouterFunc(session *api.Session, msg *api.NetworkMessage) *api.Error {
	router := routers.Get(msg.GetID())
	if router == nil {
		return api.ErrNetworkRoute
	}
	to := session.Agent
	nodeInfo := session.GetEntity().Node()
	if !slices.Contains(nodeInfo.GetTags(), router.GetNodeType()) {
		to = api.NewPidWithName(router.GetNodeType())
	}
	message := api.BuildNetMessage(session, router.GetMethod(), msg)
	if err := nodeInfo.System().PostMessage(to, message); err != nil {
		return err
	}
	return nil
}

func RunGateNode(path string) {
	gateNode := node.New(path)

	producer := func() api.IActor { return new(ServerAgent) }

	routers.Add(1, &network.Router{
		NodeType: "gate",
		ActorId:  0,
		Method:   "Login",
	})
	routers.Add(2, &network.Router{
		NodeType: "chat",
		ActorId:  0,
		Method:   "Chat",
	})
	addrArray := gateNode.GetViper().GetStringSlice("network")
	for _, addr := range addrArray {
		netModule := network.NewListener(gateNode, addr,
			network.WithHandshakeAuth(HandshakeAuthFunc),
			network.WithAgentProducer(producer),
			network.WithRouterHandler(NetRouterFunc))

		gateNode.AddModule(netModule)
	}

	gateNode.Run()

	gateNode.Wait()
}
