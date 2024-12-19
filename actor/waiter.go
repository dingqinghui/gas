/**
 * @Author: dingQingHui
 * @Description:
 * @File: future
 * @Version: 1.0.0
 * @Date: 2024/10/24 14:23
 */

package actor

import (
	"github.com/dingqinghui/gas/api"
	"time"
)

func newChanWaiter(timeout time.Duration) *chanWaiter {
	f := new(chanWaiter)
	f.ch = make(chan *api.RespondMessage, 1)
	f.after = time.After(timeout)
	return f
}

type chanWaiter struct {
	ch    chan *api.RespondMessage
	after <-chan time.Time
}

func (f *chanWaiter) Wait() (*api.RespondMessage, *api.Error) {
	select {
	case rsp := <-f.ch:
		return rsp, nil
	case <-f.after:
		return nil, api.ErrActorCallTimeout
	}
}

func (f *chanWaiter) Done(message *api.RespondMessage) {
	f.ch <- message
}
