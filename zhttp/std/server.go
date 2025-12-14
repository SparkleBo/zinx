package std

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SparkleBo/zinx/ziface"
	"github.com/SparkleBo/zinx/zrouter"
)

// Server 基于 net/http 的高性能可扩展服务器（MVP）
type Server struct {
    addr       string
    router     ziface.Router
    mws        []ziface.Middleware
    httpServer *http.Server
}

func New(addr string) *Server {
    return &Server{addr: addr, router: zrouter.New()}
}

// Use 注册全局中间件
func (s *Server) Use(mws ...ziface.Middleware) { s.mws = append(s.mws, mws...) }

// Route 注册路由与其专属中间件
func (s *Server) Route(method, path string, h ziface.Handler, mws ...ziface.Middleware) {
    s.router.Handle(method, path, h, mws...)
}

// Group 返回带前缀与中间件的子 Router （便于模块化）
func (s *Server) Group(prefix string, mws ...ziface.Middleware) ziface.Router {
    return s.router.Group(prefix, mws...)
}

// Start 启动 HTTP 服务器（非阻塞）
func (s *Server) Start() {
    if s.httpServer != nil {
        return
    }
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        h, params, mws, ok := s.router.Find(r.Method, r.URL.Path)
        if !ok {
            http.NotFound(w, r)
            return
        }
        ctx := AcquireContext(w, r)
        if len(params) > 0 { ctx.AttachParams(params) }
        final := chain(h, append(s.mws, mws...)...)
        if err := final(ctx); err != nil {
            _ = ctx.String(http.StatusInternalServerError, fmt.Sprintf("internal error: %v", err))
        }
        ReleaseContext(ctx)
    })

    s.httpServer = &http.Server{Addr: s.addr, Handler: handler}
    go func() {
        fmt.Printf("[HTTP] Listening on %s\n", s.addr)
        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("[ERROR] http server listen: %v\n", err)
        }
    }()
}

// Stop 优雅停止 HTTP 服务器
func (s *Server) Stop() {
    if s.httpServer == nil {
        return
    }
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := s.httpServer.Shutdown(ctx); err != nil {
        fmt.Printf("[ERROR] http server shutdown: %v\n", err)
    }
}

// Serve 启动并阻塞当前 goroutine
func (s *Server) Serve() {
    s.Start()
    select {}
}

// chain 构造中间件调用链，按注册顺序应用
func chain(h ziface.Handler, mws ...ziface.Middleware) ziface.Handler {
    if len(mws) == 0 { return h }
    final := h
    for i := len(mws) - 1; i >= 0; i-- {
        final = mws[i](final)
    }
    return final
}
