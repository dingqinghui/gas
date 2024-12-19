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
	"github.com/duke-git/lancet/v2/maputil"
	"reflect"
)

const fixedInnerArgNum = 2
const fixedNetworkArgNum = 3

var typeOfBytes = reflect.TypeOf(([]byte)(nil))

func NewRouterHub() api.IActorRouterHub {
	r := new(RouterHub)
	r.dict = maputil.NewConcurrentMap[string, api.IActorRouter](10)
	return r
}

type RouterHub struct {
	dict *maputil.ConcurrentMap[string, api.IActorRouter]
}

func (r *RouterHub) Set(name string, router api.IActorRouter) {
	r.dict.Set(name, router)
}

func (r *RouterHub) GetOrSet(actor api.IActor) api.IActorRouter {
	name := reflectx.TypeFullName(actor)
	v, ok := r.dict.Get(name)
	if !ok {
		v, ok = r.dict.GetOrSet(name, NewRouter(name, actor))
	}
	return v
}

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

type networkMethod struct {
	*reflectx.Method
}

func (m *networkMethod) NewRequest(ctx api.IActorContext, data []byte) (interface{}, *api.Error) {
	if m.IsRawRequest() {
		return data, nil
	} else {
		request := reflectx.NewByType(m.GetRequestType())
		if data == nil {
			return nil, nil
		}
		if err := ctx.System().Serializer().Unmarshal(data, request); err != nil {
			return nil, api.ErrJsonUnPack
		}
		return request, nil
	}
}

func (m *networkMethod) IsRawRequest() bool {
	return m.GetRequestType() == typeOfBytes
}

func (m *networkMethod) GetRequestType() reflect.Type {
	return m.ArgTypes[2]
}

func (m *networkMethod) call(ctx api.IActorContext, msg *api.ActorMessage) *api.Error {
	if len(m.ArgTypes) != fixedNetworkArgNum {
		return api.ErrActorMethodArgNum
	}
	argValues := make([]reflect.Value, fixedNetworkArgNum, fixedNetworkArgNum)
	argValues[0] = reflect.ValueOf(ctx.Actor())
	argValues[1] = reflect.ValueOf(msg.Session)
	request, err := m.NewRequest(ctx, msg.Data)
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

func (m *innerMethod) NewRequest(ctx api.IActorContext, data []byte) (interface{}, *api.Error) {
	if m.IsRawRequest() {
		return data, nil
	} else {
		request := reflectx.NewByType(m.GetRequestType())
		if data == nil {
			return nil, nil
		}
		if err := ctx.System().Serializer().Unmarshal(data, request); err != nil {
			return nil, api.ErrJsonUnPack
		}
		return request, nil
	}
}

func (m *innerMethod) IsRawRequest() bool {
	return m.GetRequestType() == typeOfBytes
}

func (m *innerMethod) GetRequestType() reflect.Type {
	return m.ArgTypes[1]
}

func (m *innerMethod) call(ctx api.IActorContext, msg *api.ActorMessage) (rsp *api.RespondMessage) {
	rsp = new(api.RespondMessage)
	if m.ArgNum != fixedInnerArgNum {
		rsp.Err = api.ErrActorArgsNum
		return
	}
	argValues := make([]reflect.Value, m.ArgNum, m.ArgNum)
	argValues[0] = reflect.ValueOf(ctx.Actor())

	request, err := m.NewRequest(ctx, msg.Data)
	if err != nil {
		rsp.Err = err
		return
	}
	argValues[1] = reflect.ValueOf(request)
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
		if !values[0].IsNil() {
			errInter := values[1].Interface()
			if errInter != nil {
				rsp.Err = errInter.(*api.Error)
			}
		}
	}
}

func (m *innerMethod) HasReply() bool {
	return m.ArgNum >= fixedInnerArgNum+1
}
