/**
 * @Author: dingQingHui
 * @Description:
 * @File: route
 * @Version: 1.0.0
 * @Date: 2024/11/5 15:24
 */

package actor

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/reflectx"
	"reflect"
)

const fixedInnerArgNum = 1
const fixedNetworkArgNum = 3

var typeOfBytes = reflect.TypeOf(([]byte)(nil))

func NewRouter(name string, actor api.IActor) *Router {
	r := new(Router)
	r.dict = reflectx.SuitableMethods(actor)
	r.name = name
	return r
}

type Router struct {
	name string
	dict map[string]*reflectx.Method
}

func (h *Router) Get(name string) *reflectx.Method {
	v, _ := h.dict[name]
	return v
}

func (h *Router) Set(name string, method *reflectx.Method) {
	h.dict[name] = method
}

func newArg(argType reflect.Type, serializer api.ISerializer, data []byte) (arg any, err *api.Error) {
	if argType == typeOfBytes {
		arg = data
		return
	}
	arg = reflectx.NewByType(argType)
	if data == nil {
		return
	}
	if e := serializer.Unmarshal(data, arg); e != nil {
		err = api.ErrUnmarshal
		return
	}
	return
}

type networkMethod struct {
	*reflectx.Method
}

func (m *networkMethod) call(ctx api.IActorContext, msg *api.ActorMessage) *api.Error {
	if len(m.ArgTypes) != fixedNetworkArgNum {
		return api.ErrActorMethodArgNum
	}
	argValues := make([]reflect.Value, fixedNetworkArgNum, fixedNetworkArgNum)
	argValues[0] = reflect.ValueOf(ctx.Actor())
	argValues[1] = reflect.ValueOf(msg.Session)
	request, err := newArg(m.ArgTypes[2], ctx.System().Serializer(), msg.Data)
	if err != nil {
		return err
	}
	argValues[2] = reflect.ValueOf(request)
	returnValues := m.Fun.Call(argValues)
	if returnValues == nil {
		return nil
	}
	errInter := returnValues[0].Interface()
	if errInter != nil {
		return errInter.(*api.Error)
	}
	return nil
}

type innerMethod struct {
	*reflectx.Method
}

func (m *innerMethod) call(ctx api.IActorContext, msg *api.ActorMessage) (rsp *api.RespondMessage) {
	rsp = new(api.RespondMessage)
	if m.ArgNum < fixedInnerArgNum {
		rsp.Err = api.ErrActorArgsNum
		return
	}
	argValues := make([]reflect.Value, m.ArgNum, m.ArgNum)
	argValues[0] = reflect.ValueOf(ctx.Actor())
	if m.ArgNum >= fixedInnerArgNum+1 {
		request, err := newArg(m.ArgTypes[1], ctx.System().Serializer(), msg.Data)
		if err != nil {
			rsp.Err = err
			return
		}
		argValues[1] = reflect.ValueOf(request)
	}
	values := m.Fun.Call(argValues)
	m.returnValues(ctx, values, rsp)
	return
}

func (m *innerMethod) returnValues(ctx api.IActorContext, values []reflect.Value, rsp *api.RespondMessage) {
	if values == nil {
		return
	}
	returnCnt := len(values)
	switch returnCnt {
	case 1:
		if !values[0].IsNil() {
			errInter := values[0].Interface()
			if errInter != nil {
				rsp.Err = errInter.(*api.Error)
			}
		}
	case 2:
		if !values[0].IsNil() {
			if buf, err := ctx.System().Serializer().Marshal(values[0].Interface()); err != nil {
				rsp.Err = api.ErrMarshal
				return
			} else {
				rsp.Data = buf
			}
		}
		if !values[1].IsNil() {
			errInter := values[1].Interface()
			if errInter != nil {
				rsp.Err = errInter.(*api.Error)
			}
		}
	}
}
