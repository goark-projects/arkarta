package nethttp

import (
	"context"
	"errors"
	"net"
	"net/http"
)

const defaultServerAddress = ":8080"

// Server 将 Arkarta net/http 参考容器绑定到标准库 HTTP Server。
type Server struct {
	container  *Container
	httpServer *http.Server
}

// NewServer 创建可监听网络连接的参考服务器。
func NewServer(container *Container, options ...ServerOption) (*Server, error) {
	if container == nil {
		return nil, ErrNilContainer
	}
	httpServer := &http.Server{
		Addr:    defaultServerAddress,
		Handler: container.Handler(),
	}
	for _, option := range options {
		if option != nil {
			option(httpServer)
		}
	}
	httpServer.Handler = container.Handler()
	return &Server{
		container:  container,
		httpServer: httpServer,
	}, nil
}

// HTTPServer 返回底层标准库服务器。
func (s *Server) HTTPServer() *http.Server {
	if s == nil {
		return nil
	}
	return s.httpServer
}

// Handler 返回容器聚合后的 HTTP Handler。
func (s *Server) Handler() http.Handler {
	if s == nil || s.container == nil {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		})
	}
	return s.container.Handler()
}

// ListenAndServe 启动 TCP 监听并在 ctx 取消时优雅关闭。
func (s *Server) ListenAndServe(ctx context.Context) error {
	if s == nil || s.container == nil || s.httpServer == nil {
		return ErrNilContainer
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := s.container.Start(ctx); err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()
	return s.wait(ctx, errCh)
}

// Serve 在调用方提供的 Listener 上服务请求，并在 ctx 取消时优雅关闭。
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	if s == nil || s.container == nil || s.httpServer == nil {
		return ErrNilContainer
	}
	if listener == nil {
		return ErrNilListener
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := s.container.Start(ctx); err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.Serve(listener)
	}()
	return s.wait(ctx, errCh)
}

// Shutdown 优雅关闭 HTTP Server 并停止所有已部署应用。
func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.container == nil || s.httpServer == nil {
		return ErrNilContainer
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return errors.Join(s.httpServer.Shutdown(ctx), s.container.Shutdown(ctx))
}

func (s *Server) wait(ctx context.Context, errCh <-chan error) error {
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownErr := s.Shutdown(context.Background())
		serverErr := <-errCh
		if errors.Is(serverErr, http.ErrServerClosed) {
			return errors.Join(ctx.Err(), shutdownErr)
		}
		return errors.Join(ctx.Err(), shutdownErr, serverErr)
	}
}
