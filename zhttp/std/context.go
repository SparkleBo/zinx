package std

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/SparkleBo/zinx/ziface"
)

// StdContext 基于 net/http 的上下文实现
type StdContext struct {
    w       http.ResponseWriter
    r       *http.Request
    storage map[string]any
    params  map[string]string
}

// --- pooling ---
var ctxPool = sync.Pool{New: func() any {
    return &StdContext{storage: make(map[string]any), params: make(map[string]string)}
}}

// AcquireContext 从对象池获取并初始化请求相关字段
func AcquireContext(w http.ResponseWriter, r *http.Request) *StdContext {
    c := ctxPool.Get().(*StdContext)
    c.w = w
    c.r = r
    // storage/params 在 Release 时已经清理，可直接复用
    return c
}

// ReleaseContext 将上下文归还对象池，清理请求相关状态
func ReleaseContext(c *StdContext) {
    // 清空共享状态与参数，避免数据泄露到下一次请求
    for k := range c.storage { delete(c.storage, k) }
    for k := range c.params { delete(c.params, k) }
    c.w = nil
    c.r = nil
    ctxPool.Put(c)
}

// NewContext 为兼容旧用法，内部走对象池
func NewContext(w http.ResponseWriter, r *http.Request) *StdContext { return AcquireContext(w, r) }

// AttachParams 设置路由参数
func (c *StdContext) AttachParams(p map[string]string) {
    c.params = p
}

// Context implements ziface.Context
func (c *StdContext) Context() context.Context { return c.r.Context() }
func (c *StdContext) Deadline() (time.Time, bool) { return c.r.Context().Deadline() }
func (c *StdContext) Done() <-chan struct{} { return c.r.Context().Done() }
func (c *StdContext) Err() error { return c.r.Context().Err() }

// Request info
func (c *StdContext) Method() string { return c.r.Method }
func (c *StdContext) Path() string { return c.r.URL.Path }
func (c *StdContext) Param(name string) string { return c.params[name] }
func (c *StdContext) Query(key string) string { return c.r.URL.Query().Get(key) }

// Shared state
func (c *StdContext) Set(key string, val any) { c.storage[key] = val }
func (c *StdContext) Get(key string) (any, bool) {
    v, ok := c.storage[key]
    return v, ok
}

// Renderers
func (c *StdContext) JSON(code int, v any) error {
    c.w.Header().Set("Content-Type", "application/json; charset=utf-8")
    c.w.WriteHeader(code)
    enc := json.NewEncoder(c.w)
    return enc.Encode(v)
}

func (c *StdContext) String(code int, s string) error {
    c.w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    c.w.WriteHeader(code)
    _, err := c.w.Write([]byte(s))
    return err
}

func (c *StdContext) Bytes(code int, b []byte) error {
    c.w.Header().Set("Content-Type", "application/octet-stream")
    c.w.WriteHeader(code)
    _, err := c.w.Write(b)
    return err
}

// 确保 StdContext 满足 ziface.Context 接口（编译期断言）
var _ ziface.Context = (*StdContext)(nil)
