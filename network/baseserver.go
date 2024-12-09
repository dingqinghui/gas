/**
 * @Author: dingQingHui
 * @Description:
 * @File: baseservice
 * @Version: 1.0.0
 * @Date: 2024/11/29 14:49
 */

package network

import (
	"context"
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/netx"
	"github.com/dingqinghui/gas/extend/serializer"
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
)

type builtinCodec struct {
}

func (b *builtinCodec) Encode(msg interface{}) ([]byte, error) {
	return serializer.Json.Marshal(msg)
}
func (b *builtinCodec) Decode(data []byte) (method string, body []byte) {
	return "Data", data
}

func newBaseService(node api.INode, typ api.NetServerType, handler gnet.EventHandler, protoAddr string, option ...Option) *baseServer {
	b := new(baseServer)
	b.typ = typ
	b.protoAddr = protoAddr
	b.handler = handler
	b.opts = loadOptions(option...)
	b.SetNode(node)
	b.Init()
	return b
}

type baseServer struct {
	api.BuiltinModule
	*gnet.BuiltinEventEngine
	protoAddr   string
	proto, addr string
	opts        *Options
	eng         gnet.Engine
	handler     gnet.EventHandler
	typ         api.NetServerType
}

func (b *baseServer) Init() {
	proto, addr, err := netx.ParseProtoAddr(b.protoAddr)
	xerror.Assert(err)
	b.proto = proto
	b.addr = addr
}

func (b *baseServer) Run() {
	if b.typ == api.NetServiceListener {
		b.Node().Workers().Submit(func() {
			xerror.Assert(gnet.Run(b.handler, b.protoAddr, b.opts.GNetOpts...))
		}, nil)
	}
}

func (b *baseServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	b.eng = eng
	return gnet.None
}

func (b *baseServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	if c == nil {
		return nil, gnet.Close
	}
	pid, err := b.spawn(c)
	if err != nil {
		return nil, gnet.Close
	}
	if err = b.Node().ActorSystem().Send(nil, pid, "Open", nil); err != nil {
		return nil, gnet.Close
	}
	c.SetContext(pid)
	return nil, gnet.None
}

func (b *baseServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	pid := b.getSessionActor(c)
	if pid == nil {
		return gnet.Close
	}
	packets := packCodec.Decode(c)
	msg := &GNetTrafficMessage{
		Packets: packets,
	}
	if err := b.Node().Send(nil, pid, "Traffic", msg); err != nil {
		return gnet.Close
	}
	return gnet.None
}

func (b *baseServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	pid := b.getSessionActor(c)
	if pid == nil {
		return gnet.Close
	}
	if err = b.Node().ActorSystem().Kill(pid); err != nil {
		return gnet.Close
	}
	return gnet.None
}

func (b *baseServer) OnShutdown(eng gnet.Engine) {

}

func (b *baseServer) getSessionActor(c gnet.Conn) *api.Pid {
	if c == nil {
		return nil
	}
	ctx := c.Context()
	if ctx == nil {
		return nil
	}
	return ctx.(*api.Pid)
}

func (b *baseServer) Type() api.NetServerType {
	return b.typ
}

func (b *baseServer) spawn(rawCon gnet.Conn) (*api.Pid, error) {
	return b.Node().ActorSystem().Spawn(func() api.IActor {
		return &SessionActor{
			opts:    b.opts,
			rawCon:  rawCon,
			service: b,
			node:    b.Node(),
		}
	}, nil)
}

func (b *baseServer) Stop() error {
	if err := b.BuiltinStopper.Stop(); err != nil {
		return err
	}
	if err := b.eng.Stop(context.Background()); err != nil {
		return err
	}
	b.Log().Info("net service module stop", zap.String("protoAddr", b.protoAddr))
	return nil
}
