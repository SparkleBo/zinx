package zmw

import (
	"net/http/httptest"
	"testing"

	"github.com/SparkleBo/zinx/zhttp/std"
	"github.com/SparkleBo/zinx/ziface"
)

func TestRecovery(t *testing.T) {
    mw := Recovery()
    panicHandler := func(ctx ziface.Context) error { panic("boom") }
    h := mw(panicHandler)

    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    ctx := std.NewContext(rr, req)
    _ = h(ctx)

    if rr.Code != 500 {
        t.Fatalf("expected 500, got %d", rr.Code)
    }
}

func TestRateLimit(t *testing.T) {
    mw := RateLimit(1, 1) // 每秒 1 个令牌，容量 1
    called := 0
    base := func(ctx ziface.Context) error { called++; return ctx.String(200, "ok") }
    h := mw(base)

    // 第一次应通过
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    ctx := std.NewContext(rr, req)
    _ = h(ctx)
    if rr.Code != 200 { t.Fatalf("expected 200, got %d", rr.Code) }

    // 立即第二次应被限流（429），因为桶容量为 1 且未补充令牌
    rr2 := httptest.NewRecorder()
    req2 := httptest.NewRequest("GET", "/", nil)
    ctx2 := std.NewContext(rr2, req2)
    _ = h(ctx2)
    if rr2.Code != 429 { t.Fatalf("expected 429, got %d", rr2.Code) }
}

