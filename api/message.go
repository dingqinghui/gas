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

const (
	InitFuncName = "OnInit"
	StopFuncName = "OnStop"
)

type (
	Message struct {
		Typ     MessageEnum
		Method  string
		From    *Pid
		To      *Pid
		Data    []byte
		Session *Session
		respond RespondFun
	}
	RespondMessage struct {
		Data []byte
		Err  *Error
	}
	RespondFun func(rsp *RespondMessage) *Error
)

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
