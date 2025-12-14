package zmw

import (
	"log"
	"time"

	"github.com/SparkleBo/zinx/ziface"
)

// Logging 访问日志中间件：记录方法、路径、耗时
func Logging() ziface.Middleware {
    return func(next ziface.Handler) ziface.Handler {
        return func(ctx ziface.Context) error {
            start := time.Now()
            err := next(ctx)
            dur := time.Since(start)
            log.Printf("%s %s %v", ctx.Method(), ctx.Path(), dur)
            return err
        }
    }
}

