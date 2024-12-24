/**
 * @Author: dingQingHui
 * @Description:
 * @File: pool
 * @Version: 1.0.0
 * @Date: 2024/8/15 15:04
 */

package workers

import (
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/dingqinghui/gas/zlog"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"sync/atomic"
)

func New() *workers {
	w := new(workers)
	pool, err := ants.NewPool(1000)
	xerror.Assert(err)
	w.pool = pool
	return w
}

type workers struct {
	pool       *ants.Pool
	count      atomic.Int64
	panicCount atomic.Uint64
}

func (w *workers) Submit(fn func(), recoverFun func(err interface{})) {

	err := w.pool.Submit(func() {
		w.count.Add(1)
		w.Try(fn, recoverFun)
		w.count.Add(-1)
	})
	if err != nil {
		return
	}
}

func (w *workers) Try(fn func(), reFun func(err interface{})) {
	defer func() {
		if err := recover(); err != nil {
			w.panicCount.Add(1)
			if reFun != nil {
				reFun(err)
			}
			zlog.Error("panic", zap.Error(err.(error)), zap.Stack("stack"))
		}
	}()
	fn()
}
