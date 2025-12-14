package std

import (
    "strings"

    "github.com/SparkleBo/zinx/ziface"
)

type segment struct {
    literal   string
    paramName string
    wildcard  bool
}

type route struct {
    method   string
    segments []segment
    handler  ziface.Handler
    mws      []ziface.Middleware
}

// Router 简易路由器，支持 :param 与 * 通配符（匹配余下所有段）
type Router struct {
    prefix string
    mws    []ziface.Middleware
    routes []*route
}

func NewRouter() *Router { return &Router{} }

func (r *Router) Handle(method, path string, h ziface.Handler, mws ...ziface.Middleware) {
    full := joinPath(r.prefix, path)
    segs := parseSegments(full)
    rr := &route{method: strings.ToUpper(method), segments: segs, handler: h, mws: append(r.mws, mws...)}
    r.routes = append(r.routes, rr)
}

// Group 返回子路由器，所有注册最终写入父路由器，确保统一匹配
func (r *Router) Group(prefix string, mws ...ziface.Middleware) ziface.Router {
    return &subRouter{parent: r, prefix: joinPath(r.prefix, prefix), mws: append(append([]ziface.Middleware{}, r.mws...), mws...)}
}

// 匹配路由并返回参数
func (r *Router) match(method, path string) (*route, map[string]string, bool) {
    method = strings.ToUpper(method)
    reqSegs := parseSegments(path)
    for i := len(r.routes) - 1; i >= 0; i-- { // 后注册优先（便于覆盖）
        rt := r.routes[i]
        if rt.method != method {
            continue
        }
        params := map[string]string{}
        if matchSegments(rt.segments, reqSegs, params) {
            return rt, params, true
        }
    }
    return nil, nil, false
}

func parseSegments(path string) []segment {
    if path == "/" { return []segment{} }
    p := strings.Trim(path, "/")
    if p == "" { return []segment{} }
    parts := strings.Split(p, "/")
    segs := make([]segment, 0, len(parts))
    for _, part := range parts {
        s := segment{}
        if part == "*" {
            s.wildcard = true
        } else if strings.HasPrefix(part, ":") {
            s.paramName = part[1:]
        } else {
            s.literal = part
        }
        segs = append(segs, s)
    }
    return segs
}

func matchSegments(rule, req []segment, params map[string]string) bool {
    ri, qi := 0, 0
    for ri < len(rule) && qi < len(req) {
        rseg := rule[ri]
        qseg := req[qi]
        if rseg.wildcard {
            // * 匹配余下所有段
            return true
        }
        if rseg.literal != "" {
            if rseg.literal != qseg.literal {
                return false
            }
        } else if rseg.paramName != "" {
            // 参数匹配
            // 注意：参数值取原始路径段（literal 字段中存原始段）
            params[rseg.paramName] = qseg.literal
        }
        ri++
        qi++
    }
    // 如果 rule 还有剩余，只有在最后一个是 wildcard 时可以匹配
    if ri < len(rule) {
        if ri == len(rule)-1 && rule[ri].wildcard {
            return true
        }
        return false
    }
    // req 还有剩余，只有 rule 最后一个是 wildcard 时允许
    if qi < len(req) {
        if len(rule) > 0 && rule[len(rule)-1].wildcard {
            return true
        }
        return false
    }
    return true
}

func joinPath(a, b string) string {
    if a == "" || a == "/" { a = "" }
    if b == "" || b == "/" { b = "" }
    if a == "" && b == "" { return "/" }
    if a == "" { return ensureSlashPrefix(b) }
    if b == "" { return ensureSlashPrefix(a) }
    return ensureSlashPrefix(strings.TrimRight(a, "/") + "/" + strings.TrimLeft(b, "/"))
}

func ensureSlashPrefix(p string) string {
    if strings.HasPrefix(p, "/") { return p }
    return "/" + p
}

// 编译期断言：Router 满足 ziface.Router
var _ ziface.Router = (*Router)(nil)

// 子路由器：转发注册到父路由器，保证路由存储集中
type subRouter struct {
    parent *Router
    prefix string
    mws    []ziface.Middleware
}

func (g *subRouter) Handle(method, path string, h ziface.Handler, mws ...ziface.Middleware) {
    // 将前缀与中间件合并后注册到父路由器
    g.parent.Handle(method, joinPath(g.prefix, path), h, append(g.mws, mws...)...)
}

func (g *subRouter) Group(prefix string, mws ...ziface.Middleware) ziface.Router {
    return &subRouter{parent: g.parent, prefix: joinPath(g.prefix, prefix), mws: append(append([]ziface.Middleware{}, g.mws...), mws...)}
}

var _ ziface.Router = (*subRouter)(nil)
