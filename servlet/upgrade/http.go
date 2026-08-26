package upgrade

import (
	"context"

	"goark.dev/arkarta/servlet"
)

// HTTPUpgrader 表示容器支持 HTTP 协议升级。
type HTTPUpgrader interface {
	UpgradeHTTP(ctx context.Context, req *servlet.Request, handler Handler) error
}

// HTTP 将当前请求升级并移交连接所有权。
func HTTP(ctx context.Context, req *servlet.Request, res servlet.Response, handler Handler) error {
	if handler == nil {
		return ErrNilHandler
	}
	if res == nil {
		return servlet.ErrNilResponse
	}
	if res.Committed() {
		return ErrAlreadyCommitted
	}
	upgrader, ok := res.(HTTPUpgrader)
	if !ok {
		return ErrUnsupported
	}
	return upgrader.UpgradeHTTP(ctx, req, handler)
}

// SwitchingProtocols 设置标准 101 响应头。
func SwitchingProtocols(res servlet.Response, protocol string) error {
	if err := servlet.SetHeader(res, "Connection", "Upgrade"); err != nil {
		return err
	}
	if protocol != "" {
		if err := servlet.SetHeader(res, "Upgrade", protocol); err != nil {
			return err
		}
	}
	res.SetStatus(101)
	return nil
}
