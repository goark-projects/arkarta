package servlet

import (
	"net/http"

	"goark.dev/arkarta/servlet/upgrade"
	"goark.dev/arkarta/websocket"
)

// WriteHandshakeResponse 向升级后的连接写出标准 HTTP 101 握手响应。
func WriteHandshakeResponse(conn upgrade.Connection, handshake websocket.Handshake) error {
	if conn == nil {
		return ErrNilConnection
	}
	response := &http.Response{
		StatusCode: http.StatusSwitchingProtocols,
		Status:     "101 " + http.StatusText(http.StatusSwitchingProtocols),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     handshake.Header(),
	}
	if buffered, ok := conn.(upgrade.BufferedConnection); ok && buffered.Writer() != nil {
		writer := buffered.Writer()
		if err := response.Write(writer); err != nil {
			return err
		}
		return writer.Flush()
	}
	return response.Write(conn)
}
