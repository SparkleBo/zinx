package ziface

import (
    "context"
    "time"
)

// Context 抽象统一请求上下文，屏蔽具体协议差异
type Context interface {
    // 基础
    Context() context.Context
    Deadline() (time.Time, bool)
    Done() <-chan struct{}
    Err() error

    // 请求信息
    Method() string
    Path() string
    Param(name string) string
    Query(key string) string

    // 共享状态
    Set(key string, val any)
    Get(key string) (any, bool)

    // 输出
    JSON(code int, v any) error
    String(code int, s string) error
    Bytes(code int, b []byte) error
}

