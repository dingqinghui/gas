/**
 * @Author: dingQingHui
 * @Description:
 * @File: Session
 * @Version: 1.0.0
 * @Date: 2024/12/16 10:46
 */

package api

func NewSession(entity INetEntity, msg *NetworkMessage) *Session {
	ses := new(Session)
	ses.Msg = msg
	ses.Agent = entity.GetAgent()
	ses.entity = entity
	return ses
}

type Session struct {
	Msg    *NetworkMessage
	Agent  *Pid
	entity INetEntity
}

func (s *Session) GetEntity() INetEntity {
	return s.entity
}
