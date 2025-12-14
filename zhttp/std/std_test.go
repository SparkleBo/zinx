package std

import (
	"net/http/httptest"
	"testing"

	"github.com/SparkleBo/zinx/ziface"
)

// --- Router unit tests ---

func TestRouter_ParamMatch(t *testing.T) {
    r := NewRouter()
    called := false
    r.Handle("GET", "/users/:id", func(ctx ziface.Context) error {
        called = true
        if ctx.Param("id") != "123" {
            t.Fatalf("param id mismatch: %s", ctx.Param("id"))
        }
        return nil
    })

    h, params, mws, ok := r.Find("GET", "/users/123")
    if !ok { t.Fatalf("should match route") }
    if params["id"] != "123" { t.Fatalf("param id expected 123, got %s", params["id"]) }
    // simulate handler call via chain
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/users/123", nil)
    ctx := AcquireContext(rr, req)
    ctx.AttachParams(params)
    final := chain(h, mws...)
    _ = final(ctx)
    ReleaseContext(ctx)
    if !called {
        t.Fatalf("handler not called")
    }
}

func TestRouter_Wildcard(t *testing.T) {
    r := NewRouter()
    r.Handle("GET", "/static/*", func(ctx ziface.Context) error { return ctx.String(200, "ok") })
    h, params, mws, ok := r.Find("GET", "/static/css/app.css")
    if !ok || h == nil { t.Fatalf("wildcard route should match") }
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/static/css/app.css", nil)
    ctx := AcquireContext(rr, req)
    ctx.AttachParams(params)
    final := chain(h, mws...)
    _ = final(ctx)
    ReleaseContext(ctx)
    if rr.Code != 200 {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
}

func TestRouter_Group(t *testing.T) {
    r := NewRouter()
    g := r.Group("/api")
    g.Handle("GET", "/v1/ping", func(ctx ziface.Context) error { return ctx.String(200, "pong") })
    h, _, _, ok := r.Find("GET", "/api/v1/ping")
    if !ok || h == nil { t.Fatalf("group route should match via parent router") }
}

// --- Context unit tests ---

func TestContext_Renderers(t *testing.T) {
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    ctx := NewContext(rr, req)

    if err := ctx.JSON(201, map[string]any{"a": 1}); err != nil { t.Fatal(err) }
    if rr.Code != 201 { t.Fatalf("json code: expected 201, got %d", rr.Code) }

    rr = httptest.NewRecorder()
    ctx = NewContext(rr, req)
    if err := ctx.String(202, "hello"); err != nil { t.Fatal(err) }
    if rr.Code != 202 { t.Fatalf("string code: expected 202, got %d", rr.Code) }

    rr = httptest.NewRecorder()
    ctx = NewContext(rr, req)
    if err := ctx.Bytes(203, []byte("data")); err != nil { t.Fatal(err) }
    if rr.Code != 203 { t.Fatalf("bytes code: expected 203, got %d", rr.Code) }
}

// --- Baseline benchmarks ---

func BenchmarkRouting_Static(b *testing.B) {
    r := NewRouter()
    r.Handle("GET", "/", func(ctx ziface.Context) error { return ctx.Bytes(200, []byte("ok")) })
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    h, params, mws, ok := r.Find("GET", "/")
    if !ok { b.Fatal("no match") }
    ctx := AcquireContext(rr, req)
    ctx.AttachParams(params)
    final := chain(h, mws...)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rr = httptest.NewRecorder()
        ctx.w = rr
        _ = final(ctx)
    }
    ReleaseContext(ctx)
}

func BenchmarkRouting_Param(b *testing.B) {
    r := NewRouter()
    r.Handle("GET", "/users/:id", func(ctx ziface.Context) error { return ctx.Bytes(200, []byte(ctx.Param("id"))) })
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/users/123", nil)
    h, params, mws, ok := r.Find("GET", "/users/123")
    if !ok { b.Fatal("no match") }
    ctx := AcquireContext(rr, req)
    ctx.AttachParams(params)
    final := chain(h, mws...)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rr = httptest.NewRecorder()
        ctx.w = rr
        _ = final(ctx)
    }
    ReleaseContext(ctx)
}
