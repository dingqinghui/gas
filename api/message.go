/**
 * @Author: dingQingHui
 * @Description:
 * @File: message
 * @Version: 1.0.0
 * @Date: 2024/12/27 11:17
 */

package api

type MessageEnum int32

const (
	MessageEnumInner     MessageEnum = 0
	MessageEnumNetwork   MessageEnum = 1
	MessageEnumBroadcast MessageEnum = 2
)

type RespondMessage struct {
	Data []byte
	Err  *Error
}
type RespondFun func(rsp *RespondMessage) *Error

type Message struct {
	Typ     MessageEnum
	Method  string
	From    *Pid
	To      *Pid
	Data    []byte
	Session *Session
	respond RespondFun
}

func (m *Message) Respond(rsp *RespondMessage) *Error {
	if m.respond == nil {
		return nil
	}
	return m.respond(rsp)
}

func (m *Message) SetRespond(respond RespondFun) {
	m.respond = respond
}
func (m *Message) IsBroadcast() bool {
	return m.Typ == MessageEnumBroadcast
}

func BuildInnerMessage(from, to *Pid, methodName string, data []byte) *Message {
	return &Message{
		From:   from,
		To:     to,
		Typ:    MessageEnumInner,
		Method: methodName,
		Data:   data,
	}
}

const (
	InitFuncName = "OnInit"
	StopFuncName = "OnStop"
)

func BuildInitMessage() *Message {
	return &Message{
		Method: InitFuncName,
	}
}

func BuildStopMessage() *Message {
	return &Message{
		Method: StopFuncName,
	}
}

//func BuildBroadMessage(from *pb.Pid, service string, methodName string, data []byte) *Message {
//	return &Message{
//		From:       from,
//		To:         NewRemotePid(0, service),
//		Typ:        ActorInnerMessage,
//		MethodName: methodName,
//		Data:       data,
//	}
//}

func NewNetworkMessage(Id uint16, Data []byte) *NetworkMessage {
	return &NetworkMessage{
		Id:   Id,
		Data: Data,
	}
}

func NewErrCodeMessage(Id uint16, err *Error) *NetworkMessage {
	return &NetworkMessage{
		Id:      Id,
		ErrCode: err.Id,
	}
}
