package std

import (
	"github.com/SparkleBo/zinx/ziface"
	"github.com/SparkleBo/zinx/zrouter"
)

// Router 作为 zrouter 的薄包装，保留构造函数以兼容已有用法
type Router struct {
    inner *zrouter.Router
}

func NewRouter() *Router { return &Router{inner: zrouter.New()} }

func (r *Router) Handle(method, path string, h ziface.Handler, mws ...ziface.Middleware) {
    r.inner.Handle(method, path, h, mws...)
}

func (r *Router) Group(prefix string, mws ...ziface.Middleware) ziface.Router {
    return r.inner.Group(prefix, mws...)
}

func (r *Router) Find(method, path string) (ziface.Handler, map[string]string, []ziface.Middleware, bool) {
    return r.inner.Find(method, path)
}

var _ ziface.Router = (*Router)(nil)
