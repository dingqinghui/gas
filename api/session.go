/**
 * @Author: dingQingHui
 * @Description:
 * @File: Session
 * @Version: 1.0.0
 * @Date: 2024/12/16 10:46
 */

package api

func NewSession(node INode, entity INetEntity) *Session {
	ses := new(Session)
	ses.Sid = node.NextId()
	ses.entity = entity
	ses.Meta = make(map[string]string)
	return ses
}

type Session struct {
	Sid    int64
	Mid    uint16
	Agent  *Pid
	Meta   map[string]string
	entity INetEntity
}

func (s *Session) GetSid() int64 {
	return s.Sid
}

func (s *Session) SetMeta(key string, value string) {
	s.Meta[key] = value
}

func (s *Session) GetMeta(key string) string {
	return s.Meta[key]
}
func (s *Session) GetEntity() INetEntity {
	return s.entity
}
