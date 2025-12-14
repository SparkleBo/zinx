package ziface

// Router  Router 抽象，屏蔽具体路由树实现
type Router interface {
    Handle(method, path string, h Handler, mws ...Middleware)
    Group(prefix string, mws ...Middleware) Router
    // Find 根据方法与路径解析到处理器、参数与中间件
    Find(method, path string) (Handler, map[string]string, []Middleware, bool)
}
