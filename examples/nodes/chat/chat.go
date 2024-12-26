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
	"github.com/dingqinghui/gas/zlog"
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

func (c *Service) OnInit(ctx api.IActorContext) *api.Error {
	_ = c.BuiltinActor.OnInit(ctx)
	c.dict = make(map[int32]*api.Pid)
	return nil
}

// async handler
func (c *Service) Join(message *common.RpcRoomJoin) *api.Error {
	zlog.Info("ChatService Join", zap.Any("message", message), zap.Any("message", c.Ctx.Message().From))
	c.dict[message.UserId] = c.Ctx.Message().From
	return nil
}

// sync handler
func (c *Service) SyncJoin1(request *common.RpcRoomJoin) (*common.RpcRoomJoin, *api.Error) {
	zlog.Info("ChatService SyncJoin1", zap.Any("message", request), zap.Any("message", c.Ctx.Message().From))
	return &common.RpcRoomJoin{UserId: 12}, api.Ok
}

// network handler
func (c *Service) Chat(session *api.Session, message *common.ClientMessage) *api.Error {
	zlog.Info("ChatService Chat", zap.Any("message", message), zap.Any("message", c.Ctx.Message().From))
	message.Content = "111111111111111"
	// push to client
	if err := c.Ctx.Push(session, 2, message); err != nil {
		return err
	}
	// respond to client
	if err := c.Ctx.Response(session, message); err != nil {
		return err
	}
	return nil
}

func RunChatNode(path string) {
	chatNode := node.New(path)

	chatNode.Run()
	_, _ = chatNode.System().Spawn(func() api.IActor { return new(Service) }, api.WithActorName("chat"))
	chatNode.Wait()
}
