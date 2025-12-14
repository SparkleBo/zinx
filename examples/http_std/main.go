package main

import (
	"time"

	"github.com/SparkleBo/zinx/zhttp/std"
	"github.com/SparkleBo/zinx/ziface"
	"github.com/SparkleBo/zinx/zmw"
)

func main() {
    s := std.New(":8080")
    // 全局中间件：访问日志、恢复、限流（100 rps，burst 200）
    s.Use(zmw.Logging(), zmw.Recovery(), zmw.RateLimit(100, 200))

    // 路由：GET /
    s.Route("GET", "/", func(ctx ziface.Context) error {
        return ctx.String(200, "hello, zinx std http")
    })

    // 路由：GET /users/:id
    s.Route("GET", "/users/:id", func(ctx ziface.Context) error {
        id := ctx.Param("id")
        return ctx.JSON(200, map[string]any{"id": id, "time": time.Now().Format(time.RFC3339)})
    })

    s.Serve()
}
