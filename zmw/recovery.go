package zmw

import (
    "fmt"
    "log"
    "runtime/debug"

    "github.com/SparkleBo/zinx/ziface"
)

// Recovery 捕获 panic，输出 500 并打印堆栈
func Recovery() ziface.Middleware {
    return func(next ziface.Handler) ziface.Handler {
        return func(ctx ziface.Context) (err error) {
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("panic: %v\n%s", r, string(debug.Stack()))
                    err = ctx.String(500, fmt.Sprintf("internal error: %v", r))
                }
            }()
            return next(ctx)
        }
    }
}

