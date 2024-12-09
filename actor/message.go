/**
 * @Author: dingQingHui
 * @Description:
 * @File: system_message
 * @Version: 1.0.0
 * @Date: 2023/12/8 16:13
 */

package actor

import "github.com/dingqinghui/gas/api"

const (
	InitFuncName = "OnInit"
	StopFuncName = "OnStop"
)

var _ api.IMailBoxMessage = (*mailBoxMessage)(nil)

type mailBoxMessage struct {
	from     *api.Pid
	args     []interface{}
	funcName string
	waiter   api.IActorWaiter
}

func (e *mailBoxMessage) MethodName() string {
	return e.funcName
}

func (e *mailBoxMessage) Args() []interface{} {
	return e.args
}
func (e *mailBoxMessage) Waiter() api.IActorWaiter {
	return e.waiter
}
func (e *mailBoxMessage) From() *api.Pid {
	return e.from
}

func NewMailBoxMessage(from *api.Pid, funcName string, waiter api.IActorWaiter, args ...interface{}) *mailBoxMessage {
	return &mailBoxMessage{
		from:     from,
		args:     args,
		waiter:   waiter,
		funcName: funcName,
	}
}
