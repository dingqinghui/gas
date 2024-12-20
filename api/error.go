/**
 * @Author: dingQingHui
 * @Description:
 * @File: error
 * @Version: 1.0.0
 * @Date: 2024/10/24 10:34
 */

package api

import (
	"fmt"
	"sync"
)

type Error struct {
	Id  uint16
	Str string
}

func (r *Error) Error() string {
	return r.Str
}

var idErrMap = sync.Map{}
var errIdMap = sync.Map{}

func NewErr(str string, id uint16) *Error {
	err := &Error{id, str}
	_, ok := idErrMap.Load(id)
	if ok {
		panic(fmt.Sprintf("repeated error id:%v", id))
	}
	idErrMap.Store(id, err)
	errIdMap.Store(err, id)
	return err
}

var (
	Ok                        = NewErr("正确", 0)
	ErrMsgPackPack            = NewErr("msgpack打包错误", 1)
	ErrMsgPackUnPack          = NewErr("msgpack解析错误", 2)
	ErrPBPack                 = NewErr("pb打包错误", 3)
	ErrPBUnPack               = NewErr("pb解析错误", 4)
	ErrJsonPack               = NewErr("json打包错误", 5)
	ErrJsonUnPack             = NewErr("json解析错误", 6)
	ErrActorArgsNum           = NewErr("actor参数数量错误", 7)
	ErrActorNameExist         = NewErr("actor name exist", 8)
	ErrActorNameNotExist      = NewErr("actor name not exist", 9)
	ErrInvalidPid             = NewErr("pid invalid", 10)
	ErrNotLocalPid            = NewErr("not local pid", 11)
	ErrProcessNotExist        = NewErr("process not exist", 12)
	ErrMailBoxNil             = NewErr("mailbox is nil", 13)
	ErrActorStopped           = NewErr("actor is stopped", 14)
	ErrActorCallTimeout       = NewErr("actor call timeout", 15)
	ErrActorNotMethod         = NewErr("actor not method", 16)
	ErrActorMethodArgNum      = NewErr("actor method args num", 17)
	ErrStopped                = NewErr("stopped", 18)
	ErrDiscoveryProviderIsNil = NewErr("discovery provider is nil", 19)
	ErrNetEntityIsNil         = NewErr("net entity is nil", 20)
	ErrConsul                 = NewErr("consul err", 21)
	ErrMarshal                = NewErr("marshal err", 22)
	ErrUnmarshal              = NewErr("unmarshal err", 23)
	ErrNatsSend               = NewErr("nats send err", 24)
	ErrPidIsNil               = NewErr("pid is nil", 25)
	ErrPacketType             = NewErr("packet type err", 26)
	ErrNetworkRoute           = NewErr("network route err", 27)
	ErrGNetRaw                = NewErr("gnet raw err", 28)
	ErrNetworkRespond         = NewErr("network respond err", 29)
	ErrNatsRespond            = NewErr("nats respond err", 30)
	ErrActorRouterIsNil       = NewErr("actor router is nil", 31)
)

func IsOk(err *Error) bool {
	return err == nil || err.Id == 0
}

func Assert(err *Error) {
	if IsOk(err) {
		return
	}
	panic(err.Error())
}
