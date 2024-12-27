/**
 * @Author: dingQingHui
 * @Description:
 * @File: message
 * @Version: 1.0.0
 * @Date: 2024/12/27 11:17
 */

package api

type RespondMessage struct {
	Data []byte
	Err  *Error
}
type RespondFun func(rsp *RespondMessage) *Error

type Message struct {
	Typ        int
	MethodName string
	Session    *Session
	From       *Pid
	To         *Pid
	Data       []byte
	respond    RespondFun
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
	return m.Typ == ActorBroadcastMessage
}

func BuildNetMessage(session *Session, methodName string) *Message {
	return &Message{
		Typ:        ActorNetMessage,
		MethodName: methodName,
		Session:    session,
	}
}

func BuildInnerMessage(from, to *Pid, methodName string, data []byte) *Message {
	return &Message{
		From:       from,
		To:         to,
		Typ:        ActorInnerMessage,
		MethodName: methodName,
		Data:       data,
	}
}

func BuildBroadMessage(from *Pid, service string, methodName string, data []byte) *Message {
	return &Message{
		From:       from,
		To:         NewRemotePid(0, service),
		Typ:        ActorInnerMessage,
		MethodName: methodName,
		Data:       data,
	}
}

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
