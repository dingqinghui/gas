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
	"github.com/dingqinghui/gas/zlog"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"time"
)

func New() api.IRpcMessageQue {
	c := new(Conn)
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
	c.cfg = initConfig()
	if c.cfg == nil {
		return
	}
	c.msgChan = make(chan *nats.Msg, c.cfg.recChanSize)
	c.connect()
}

func (c *Conn) connect() {
	if c.cfg == nil {
		return
	}
	con, err := nats.Connect(c.cfg.urls)
	if err != nil {
		zlog.Error("nats connect err", zap.String("address", c.cfg.urls), zap.Error(err))
		return
	}
	c.rawCon = con
	zlog.Info("nats connect", zap.String("address", c.cfg.urls))
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
		zlog.Error("nats request", zap.String("subj", topic), zap.Error(err))
		return nil, err
	}
	zlog.Debug("nats request", zap.String("topic", topic))
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
		zlog.Error("nats publish error", zap.Error(err))
		return api.ErrNatsSend
	}
	zlog.Debug("nats request", zap.String("topic", topic))
	return nil
}

func (c *Conn) Subscribe(topic string, process api.RpcProcessHandler) {
	if api.GetNode() == nil {
		return
	}
	_, chanErr := c.rawCon.ChanSubscribe(topic, c.msgChan)
	if chanErr != nil {
		zlog.Error("nats chan subscribe error", zap.Error(chanErr))
		return
	}
	api.GetNode().Submit(func() {
		for msg := range c.msgChan {
			respond := func(data []byte) *api.Error {
				if err := msg.Respond(data); err != nil {
					zlog.Error("nats chan respond error", zap.Error(err))
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
		zlog.Panic("nats process panic", zap.Error(err.(error)))
	})

	zlog.Info("nats subscribe topic", zap.String("topic", topic))
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
	zlog.Info("nats module stop")
	return nil
}
