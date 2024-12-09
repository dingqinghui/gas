/**
 * @Author: dingQingHui
 * @Description:
 * @File: route
 * @Version: 1.0.0
 * @Date: 2024/11/5 15:24
 */

package actor

import (
	"errors"
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/reflectx"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/xerror"
	"go.uber.org/zap"
	"reflect"
)

const fixedArgNum = 1

func DefaultMethod(actor api.IActor) error { return nil }

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
	r.dict = reflectx.SuitableMethods(actor, DefaultMethod)
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

func (h *Router) Call(ctx api.IActorContext, mbMsg api.IMailBoxMessage) error {
	methodName := mbMsg.MethodName()
	args := mbMsg.Args()
	m, ok := h.dict[methodName]
	if !ok || m == nil {
		return xerror.New("actor %v not func:%v", h.name, methodName)
	}
	ctx.Debug("actor call", zap.String("actor", reflect.TypeOf(ctx.Actor()).String()),
		zap.String("methodName", methodName), zap.Any("args", mbMsg.Args()))
	return h.callMethod(ctx, m, args)
}

func (h *Router) callMethod(ctx api.IActorContext, m *reflectx.Method, args []interface{}) error {
	if len(args) != m.ArgNum-fixedArgNum {
		return errors.New("args count err")
	}
	argValues := make([]reflect.Value, m.ArgNum, m.ArgNum)
	argValues[0] = reflect.ValueOf(ctx.Actor())

	for i, arg := range args {
		if arg == nil {
			arg = reflectx.NewByType(m.ArgTypes[i+fixedArgNum])
		}
		argValues[i+fixedArgNum] = reflect.ValueOf(arg)
	}
	returnValues := m.Fun.Call(argValues)
	if returnValues == nil {
		return nil
	}
	errInter := returnValues[0].Interface()
	if errInter != nil {
		return errInter.(error)
	}
	return nil
}

func (h *Router) NewArgs(methodName string) (request, reply interface{}, err error) {
	method := h.Get(methodName)
	if method == nil {
		err = xerror.New("actor not method:%v", methodName)
		return
	}
	request = reflectx.NewByType(method.ArgTypes[1])
	if method.ArgNum >= fixedArgNum+2 {
		reply = reflectx.NewByType(method.ArgTypes[2])
	}
	return
}
