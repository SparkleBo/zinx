package ziface

// Handler 处理器函数
type Handler func(ctx Context) error

// Middleware 中间件，包装 Handler 实现链式拦截
type Middleware func(Handler) Handler

