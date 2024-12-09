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

func newWaiter(timeout time.Duration) api.IActorWaiter {
	f := new(Waiter)
	f.ch = make(chan struct{}, 1)
	f.after = time.After(timeout)
	return f
}

type Waiter struct {
	ch    chan struct{}
	after <-chan time.Time
}

func (f *Waiter) Wait() error {
	select {
	case _ = <-f.ch:
		return nil
	case <-f.after:
		return api.ErrActorCallTimeout
	}
}

func (f *Waiter) Done() {
	f.ch <- struct{}{}
}
