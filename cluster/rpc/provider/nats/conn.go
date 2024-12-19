/**
 * @Author: dingQingHui
 * @Description:
 * @File: Conn
 * @Version: 1.0.0
 * @Date: 2024/11/19 10:15
 */

package nats

import (
	"github.com/dingqinghui/gas/api"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"time"
)

func New(node api.INode) api.IRpcMessageQue {
	c := new(Conn)
	c.SetNode(node)
	c.Init()
	return c
}

type Conn struct {
	api.BuiltinModule
	rawCon  *nats.Conn
	msgChan chan *nats.Msg
	cfg     *config
}

func (c *Conn) Name() string {
	return "nats"
}

func (c *Conn) Init() {
	c.cfg = initConfig(c.Node())
	c.msgChan = make(chan *nats.Msg, c.cfg.recChanSize)
	c.connect()
}

func (c *Conn) connect() {
	con, err := nats.Connect(c.cfg.urls)
	if err != nil {
		c.Log().Error("nats connect err", zap.String("address", c.cfg.urls), zap.Error(err))
		return
	}
	c.rawCon = con
	c.Log().Info("nats connect", zap.String("address", c.cfg.urls))
}

// Call
// @Description: 同步发送请求
// @receiver c
// @param subj
// @param data
// @param timeout
// @return *nats.Msg
// @return error
func (c *Conn) Call(topic string, data []byte, timeout time.Duration) ([]byte, error) {
	msg, err := c.rawCon.Request(topic, data, timeout)
	if err != nil {
		c.Log().Error("nats request", zap.String("subj", topic), zap.Error(err))
		return nil, err
	}
	c.Log().Info("nats request", zap.String("topic", topic))
	return msg.Data, err
}

// Send
// @Description: 异步调用
// @receiver c
// @param subj
// @param data
// @return err
func (c *Conn) Send(topic string, data []byte) *api.Error {
	if err := c.rawCon.Publish(topic, data); err != nil {
		c.Log().Error("nats publish error", zap.Error(err))
		return api.ErrNatsSend
	}
	c.Log().Debug("nats request", zap.String("topic", topic))
	return nil
}

func (c *Conn) Subscribe(subject string, process api.RpcProcessHandler) {
	_, chanErr := c.rawCon.ChanSubscribe(subject, c.msgChan)
	if chanErr != nil {
		c.Log().Error("nats chan subscribe error", zap.Error(chanErr))
		return
	}
	c.Node().Workers().Submit(func() {
		for msg := range c.msgChan {
			respond := func(data []byte) *api.Error {
				if err := msg.Respond(data); err != nil {
					c.Log().Error("nats chan respond error", zap.Error(err))
					return api.ErrNatsRespond
				}
				return nil
			}
			if msg.Reply == "" {
				respond = nil
			}
			process(msg.Subject, msg.Data, respond)
		}

	}, func(err interface{}) {
		c.Log().Panic("nats process panic", zap.Error(err.(error)))
	})
}

func (c *Conn) Stop() *api.Error {
	if err := c.BuiltinStopper.Stop(); err != nil {
		return err
	}
	if c.msgChan != nil {
		c.msgChan = nil
	}
	if c.rawCon != nil {
		c.rawCon.Close()
	}
	c.Log().Info("nats module stop")
	return nil
}
