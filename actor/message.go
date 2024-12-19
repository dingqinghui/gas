/**
 * @Author: dingQingHui
 * @Description:
 * @File: system_message
 * @Version: 1.0.0
 * @Date: 2023/12/8 16:13
 */

package actor

import (
	"github.com/dingqinghui/gas/api"
)

const (
	InitFuncName = "OnInit"
	StopFuncName = "OnStop"
)

func buildInitMessage() *api.ActorMessage {
	return &api.ActorMessage{
		MethodName: InitFuncName,
	}
}

func buildStopMessage() *api.ActorMessage {
	return &api.ActorMessage{
		MethodName: StopFuncName,
	}
}
