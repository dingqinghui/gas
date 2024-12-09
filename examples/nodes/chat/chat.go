/**
 * @Author: dingQingHui
 * @Description:
 * @File: chat_test
 * @Version: 1.0.0
 * @Date: 2024/12/2 11:00
 */

package chat

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/examples/common"
	"github.com/dingqinghui/gas/node"
	"go.uber.org/zap"
)

type Message struct {
	Name    string
	Content string
}

type Service struct {
	api.BuiltinActor
	dict map[int32]*api.Pid
}

func (c *Service) OnInit(ctx api.IActorContext) error {
	_ = c.BuiltinActor.OnInit(ctx)
	c.dict = make(map[int32]*api.Pid)
	return nil
}
func (c *Service) Join(message *common.RpcRoomJoin) error {
	c.dict[message.UserId] = c.Ctx.Message().From()
	c.Ctx.Info("ChatService Join", zap.Any("message", message), zap.Any("message", c.Ctx.Message().From()))
	return c.Ctx.Send(c.Ctx.Message().From(), "JoinSuccess", &common.ClientMessage{})
}

//func (c *Service) Chat(ctx api.IActorContext, message *common.ClientMessage) error {
//	ctx.System().Log().Info("ChatService chat", zap.Any("message", message))
//	return nil
//}

func RunChatNode(path string) {
	chatNode := node.New(path)
	_, _ = chatNode.ActorSystem().SpawnWithName("ChatService", func() api.IActor { return new(Service) }, nil)
	chatNode.Run()
}
