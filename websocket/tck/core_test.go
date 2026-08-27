package tck_test

import (
	"testing"

	"goark.dev/arkarta/websocket"
	"goark.dev/arkarta/websocket/tck"
)

func TestRunHandshake(t *testing.T) {
	tck.RunHandshake(t, func(options ...websocket.HandshakeOption) *websocket.Handshaker {
		return websocket.NewHandshaker(options...)
	})
}

func TestRunEndpointLifecycle(t *testing.T) {
	tck.RunEndpointLifecycle(t, websocket.NewSession)
}

func TestRunCompression(t *testing.T) {
	tck.RunCompression(t, websocket.NewPerMessageDeflate)
}
