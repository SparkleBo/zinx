package zmw

import (
    "time"

    "github.com/SparkleBo/zinx/ziface"
)

// RateLimit 简易令牌桶限流
// rps: 每秒产生的令牌数，burst: 桶容量
func RateLimit(rps int, burst int) ziface.Middleware {
    if rps <= 0 { rps = 1 }
    if burst <= 0 { burst = 1 }

    tokens := make(chan struct{}, burst)
    // 预热填满桶
    for i := 0; i < burst; i++ { tokens <- struct{}{} }

    // 按固定速率补充令牌
    interval := time.Second / time.Duration(rps)
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            select {
            case tokens <- struct{}{}:
            default:
                // 桶满，丢弃令牌
            }
        }
    }()

    return func(next ziface.Handler) ziface.Handler {
        return func(ctx ziface.Context) error {
            select {
            case <-tokens:
                return next(ctx)
            default:
                return ctx.String(429, "too many requests")
            }
        }
    }
}

