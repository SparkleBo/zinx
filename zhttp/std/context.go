package std

import (
	"context"
	"encoding/json"
	"net/http"
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

func NewContext(w http.ResponseWriter, r *http.Request) *StdContext {
    return &StdContext{
        w:       w,
        r:       r,
        storage: make(map[string]any),
        params:  make(map[string]string),
    }
}

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

