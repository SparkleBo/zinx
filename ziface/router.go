package ziface

// Router 路由器抽象，屏蔽具体路由树实现
type Router interface {
    Handle(method, path string, h Handler, mws ...Middleware)
    Group(prefix string, mws ...Middleware) Router
}

