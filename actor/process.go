/**
 * @Author: dingQingHui
 * @Description:
 * @File: process
 * @Version: 1.0.0
 * @Date: 2024/10/24 10:43
 */

package actor

import (
	"github.com/dingqinghui/gas/api"
	"time"
)

func NewBaseProcess(ctx api.IActorContext, mailbox api.IActorMailbox) api.IProcess {
	process := &ProcessActor{
		mailbox: mailbox,
		ctx:     ctx,
	}
	return process
}

var _ api.IProcess = (*ProcessActor)(nil)

type ProcessActor struct {
	api.BuiltinStopper
	mailbox api.IActorMailbox
	ctx     api.IActorContext
}

func (p *ProcessActor) Pid() *api.Pid {
	return p.ctx.Self()
}

func (p *ProcessActor) valid() error {
	if p.IsStop() {
		return api.ErrActorStopped
	}
	if p.mailbox == nil {
		return api.ErrMailBoxNil
	}
	return nil
}

func (p *ProcessActor) Send(from *api.Pid, methodName string, request interface{}) error {
	if err := p.valid(); err != nil {
		return err
	}
	env := NewMailBoxMessage(from, methodName, nil, request)
	return p.mailbox.PostMessage(env)
}

func (p *ProcessActor) CallAndWait(from *api.Pid, methodName string, timeout time.Duration, request, reply any) error {
	if err := p.valid(); err != nil {
		return err
	}
	w := newWaiter(timeout)
	mbm := NewMailBoxMessage(from, methodName, w, request, reply)
	if err := p.mailbox.PostMessage(mbm); err != nil {
		return err
	}
	if err := w.Wait(); err != nil {
		return err
	}
	return nil
}
func (p *ProcessActor) Call(from *api.Pid, methodName string, timeout time.Duration, request, reply any) (api.IActorWaiter, error) {
	if err := p.valid(); err != nil {
		return nil, err
	}
	w := newWaiter(timeout)
	mbm := NewMailBoxMessage(from, methodName, w, request, reply)
	if err := p.mailbox.PostMessage(mbm); err != nil {
		return nil, err
	}
	return w, nil
}

func (p *ProcessActor) Stop() error {
	if err := p.BuiltinStopper.Stop(); err != nil {
		return err
	}
	w := newWaiter(time.Millisecond * 10)
	env := NewMailBoxMessage(nil, StopFuncName, w)
	if err := p.mailbox.PostMessage(env); err != nil {
		return err
	}
	if err := w.Wait(); err != nil {
		return err
	}
	return nil
}

func (p *ProcessActor) AsyncStop() error {
	if err := p.BuiltinStopper.Stop(); err != nil {
		return err
	}
	env := NewMailBoxMessage(nil, StopFuncName, nil)
	if err := p.mailbox.PostMessage(env); err != nil {
		return err
	}
	return nil
}

func (p *ProcessActor) Context() api.IActorContext {
	return p.ctx
}
