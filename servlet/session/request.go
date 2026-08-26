package session

import "goark.dev/arkarta/servlet"

// Current 返回当前请求已经关联的 Session。
func Current(req *servlet.Request) (Session, bool) {
	if req == nil {
		return nil, false
	}
	value, ok := req.Attribute(AttributeCurrentSession)
	if !ok {
		return nil, false
	}
	current, ok := value.(Session)
	if !ok || current == nil || !current.IsValid() {
		req.SetAttribute(AttributeCurrentSession, nil)
		return nil, false
	}
	return current, true
}
