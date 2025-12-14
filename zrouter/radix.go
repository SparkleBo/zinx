package zrouter

import (
	"strings"

	"github.com/SparkleBo/zinx/ziface"
)

type nodeKind uint8

const (
    nkStatic nodeKind = iota // 字面量段
    nkParam                  // :param
    nkWildcard               // *
)

type node struct {
    kind      nodeKind
    label     string            // 对于 static/param，存储段内容或参数名
    children  []*node           // 压缩后的子节点，按首字符或种类区分
    handler   ziface.Handler
    mws       []ziface.Middleware
}

// Router 使用按段压缩的 Radix/Trie
type Router struct {
    root   *node
    prefix string
    mws    []ziface.Middleware
}

func New() *Router { return &Router{root: &node{kind: nkStatic}} }

func (r *Router) Handle(method, path string, h ziface.Handler, mws ...ziface.Middleware) {
    // 将 method 合并进第一段以区分不同方法（避免额外维度）
    full := joinPath(r.prefix, path)
    segs := splitPath(full)
    // 在根下以 method 建子树
    cur := r.ensureChild(r.root, nkStatic, strings.ToUpper(method))
    // 逐段插入
    for i, s := range segs {
        var kind nodeKind
        var label string
        if s == "*" {
            kind = nkWildcard
        } else if strings.HasPrefix(s, ":") {
            kind = nkParam
            label = s[1:]
        } else {
            kind = nkStatic
            label = s
        }
        cur = r.ensureChild(cur, kind, label)
        // wildcard 必须是最后一段
        if kind == nkWildcard && i != len(segs)-1 {
            // 忽略后续段，* 吃掉余下路径
            break
        }
    }
    cur.handler = h
    cur.mws = append(append([]ziface.Middleware{}, r.mws...), mws...)
}

// Group 创建带前缀与中间件的子 Router 
func (r *Router) Group(prefix string, mws ...ziface.Middleware) ziface.Router {
    return &group{parent: r, prefix: joinPath(r.prefix, prefix), mws: append(append([]ziface.Middleware{}, r.mws...), mws...)}
}

// Find 根据方法与路径查找处理器与参数
func (r *Router) Find(method, path string) (ziface.Handler, map[string]string, []ziface.Middleware, bool) {
    segs := splitPath(path)
    params := map[string]string{}
    // 方法维度
    cur := r.childBy(r.root, nkStatic, strings.ToUpper(method))
    if cur == nil { return nil, nil, nil, false }
    for i := 0; i < len(segs); i++ {
        s := segs[i]
        // 先尝试静态匹配
        next := r.childBy(cur, nkStatic, s)
        if next != nil {
            cur = next
            continue
        }
        // 其次参数匹配
        next = r.childByKind(cur, nkParam)
        if next != nil {
            params[next.label] = s
            cur = next
            continue
        }
        // 最后 wildcard
        next = r.childByKind(cur, nkWildcard)
        if next != nil {
            cur = next
            break
        }
        return nil, nil, nil, false
    }
    if cur.handler == nil {
        // 路径完全匹配但无处理器，检查是否 wildcard 叶子有 handler
        wc := r.childByKind(cur, nkWildcard)
        if wc != nil && wc.handler != nil {
            cur = wc
        } else {
            return nil, nil, nil, false
        }
    }
    return cur.handler, params, cur.mws, true
}

// --- helpers ---

func (r *Router) ensureChild(n *node, kind nodeKind, label string) *node {
    // 查找是否已有可复用的子节点
    for _, c := range n.children {
        if c.kind == kind && c.label == label {
            return c
        }
    }
    c := &node{kind: kind, label: label}
    n.children = append(n.children, c)
    return c
}

func (r *Router) childBy(n *node, kind nodeKind, label string) *node {
    for _, c := range n.children {
        if c.kind == kind && c.label == label {
            return c
        }
    }
    return nil
}

func (r *Router) childByKind(n *node, kind nodeKind) *node {
    for _, c := range n.children {
        if c.kind == kind {
            return c
        }
    }
    return nil
}

func splitPath(path string) []string {
    if path == "/" { return []string{} }
    p := strings.Trim(path, "/")
    if p == "" { return []string{} }
    return strings.Split(p, "/")
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

// 组 Router ：所有注册写入父 Router 
type group struct {
    parent *Router
    prefix string
    mws    []ziface.Middleware
}

func (g *group) Handle(method, path string, h ziface.Handler, mws ...ziface.Middleware) {
    // 合并组的中间件与路由中间件，并将注册写入父 Router 
    g.parent.Handle(method, joinPath(g.prefix, path), h, append(g.mws, mws...)...)
}

func (g *group) Group(prefix string, mws ...ziface.Middleware) ziface.Router {
    return &group{parent: g.parent, prefix: joinPath(g.prefix, prefix), mws: append(append([]ziface.Middleware{}, g.mws...), mws...)}
}

// Find 仅为满足接口，正常查找应走顶层 Router ；这里做前缀拼接后转发
func (g *group) Find(method, path string) (ziface.Handler, map[string]string, []ziface.Middleware, bool) {
    return g.parent.Find(method, joinPath(g.prefix, path))
}

var _ ziface.Router = (*Router)(nil)
var _ ziface.Router = (*group)(nil)
