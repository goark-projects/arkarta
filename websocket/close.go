package websocket

// CloseCode 表示 WebSocket 关闭状态码。
type CloseCode int

const (
	CloseNormal              CloseCode = 1000
	CloseGoingAway           CloseCode = 1001
	CloseProtocolError       CloseCode = 1002
	CloseUnsupportedData     CloseCode = 1003
	CloseNoStatus            CloseCode = 1005
	CloseAbnormal            CloseCode = 1006
	CloseInvalidPayload      CloseCode = 1007
	ClosePolicyViolation     CloseCode = 1008
	CloseMessageTooBig       CloseCode = 1009
	CloseMandatoryExtension  CloseCode = 1010
	CloseUnexpectedCondition CloseCode = 1011
	CloseServiceRestart      CloseCode = 1012
	CloseTryAgainLater       CloseCode = 1013
	CloseTLSFailure          CloseCode = 1015
)

// CloseReason 描述 WebSocket 关闭原因。
type CloseReason struct {
	code   CloseCode
	reason string
}

// NewCloseReason 创建关闭原因。
func NewCloseReason(code CloseCode, reason string) CloseReason {
	if code == 0 {
		code = CloseNormal
	}
	return CloseReason{code: code, reason: reason}
}

// Code 返回关闭状态码。
func (r CloseReason) Code() CloseCode {
	if r.code == 0 {
		return CloseNormal
	}
	return r.code
}

// Reason 返回关闭文本。
func (r CloseReason) Reason() string {
	return r.reason
}
