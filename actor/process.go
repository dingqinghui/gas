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

func (p *ProcessActor) valid() *api.Error {
	if p.IsStop() {
		return api.ErrActorStopped
	}
	if p.mailbox == nil {
		return api.ErrMailBoxNil
	}
	return nil
}

func (p *ProcessActor) PostMessage(message *api.Message) *api.Error {
	if err := p.valid(); err != nil {
		return err
	}
	return p.mailbox.PostMessage(message)
}

func (p *ProcessActor) PostMessageAndWait(message *api.Message) (rsp *api.RespondMessage) {
	rsp = new(api.RespondMessage)
	if err := p.valid(); err != nil {
		rsp.Err = err
		return
	}
	waiter := newChanWaiter(p.ctx.System().Timeout())
	message.SetRespond(func(rsp *api.RespondMessage) *api.Error {
		waiter.Done(rsp)
		return nil
	})
	if err := p.PostMessage(message); err != nil {
		rsp.Err = err
		return
	}
	_rsp, err := waiter.Wait()
	if !api.IsOk(err) {
		rsp.Err = err
		return
	}
	if _rsp == nil {
		return
	}
	rsp = _rsp
	return
}

func (p *ProcessActor) Stop() *api.Error {
	if err := p.BuiltinStopper.Stop(); err != nil {
		return err
	}
	rsp := p.PostMessageAndWait(buildStopMessage())
	if !api.IsOk(rsp.Err) {
		return rsp.Err
	}
	return nil
}

func (p *ProcessActor) Context() api.IActorContext {
	return p.ctx
}
