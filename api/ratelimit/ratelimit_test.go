package ratelimit

import (
	"context"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeValkey implements ValkeyLike for tests.
type fakeValkey struct {
	counts       map[string]int64
	errOnIncr    error
	expireCalls  map[string]time.Duration
}

func newFakeValkey() *fakeValkey {
	return &fakeValkey{
		counts:      make(map[string]int64),
		expireCalls: make(map[string]time.Duration),
	}
}

func (f *fakeValkey) Increment(key string, ctx context.Context) (int64, error) {
	if f.errOnIncr != nil {
		return 0, f.errOnIncr
	}
	f.counts[key]++
	return f.counts[key], nil
}

func (f *fakeValkey) Expirer(key string, ctx context.Context, expiration time.Duration) (bool, error) {
	f.expireCalls[key] = expiration
	return true, nil
}

func (f *fakeValkey) ExpireNX(key string, ctx context.Context, expiration time.Duration) (bool, error) {
	if _, alreadySet := f.expireCalls[key]; alreadySet {
		return false, nil
	}
	f.expireCalls[key] = expiration
	return true, nil
}

// fakeKeyStore implements KeyStoreLike for tests.
type fakeKeyStore struct {
	limits map[[32]byte]int
}

func (f *fakeKeyStore) GetKeyLimit(hash [32]byte) (int, bool) {
	v, ok := f.limits[hash]
	return v, ok
}

// helper to build a simple next handler that just returns 200
func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestMiddleware_IPBased_AllowsAndSetsHeaders(t *testing.T) {
	vk := newFakeValkey()
	ks := &fakeKeyStore{limits: map[[32]byte]int{}}
	ipLimit := 3
	window := 30 * time.Second
	rl := NewRateLimiter(vk, ks, ipLimit, window, nil)

	mw := rl.Middleware(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Use RemoteAddr to be picked when no XFF/X-Real-Ip
	req.RemoteAddr = "203.0.113.5:12345"
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "3", res.Header.Get("X-RateLimit-Limit"))
	assert.Equal(t, "2", res.Header.Get("X-RateLimit-Remaining"))

	// ExpireNX is called on every request; verify the key received a TTL.
	key := "rl:ip:203.0.113.5"
	if _, ok := vk.expireCalls[key]; !ok {
		t.Fatalf("expected expirer to be called for key %s", key)
	}
	assert.Equal(t, window, vk.expireCalls[key])
}

func TestMiddleware_InvalidAPIKey_Unauthorized(t *testing.T) {
	vk := newFakeValkey()
	ks := &fakeKeyStore{limits: map[[32]byte]int{}}
	rl := NewRateLimiter(vk, ks, 10, time.Minute, nil)

	mw := rl.Middleware(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "unknown-key")
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddleware_ValidAPIKey_UsesKeyLimitAndHeaders(t *testing.T) {
	vk := newFakeValkey()
	// Prepare keystore with a key limit
	key := "my-api-key"
	h := sha256.Sum256([]byte(key))
	ks := &fakeKeyStore{limits: map[[32]byte]int{h: 5}}
	rl := NewRateLimiter(vk, ks, 1, time.Minute, nil)

	mw := rl.Middleware(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", key)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	res := rec.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "5", res.Header.Get("X-RateLimit-Limit"))
	assert.Equal(t, "4", res.Header.Get("X-RateLimit-Remaining"))

	// ensure key-based identifier used (first 8 bytes of hash)
	short := (h[:8])
	id := "key:" + (func() string { return (func(b []byte) string {
		const hextable = "0123456789abcdef"
		out := make([]byte, len(b)*2)
		for i, v := range b {
			out[i*2] = hextable[v>>4]
			out[i*2+1] = hextable[v&0x0f]
		}
		return string(out)
	})(short) })()
	_, ok := vk.expireCalls["rl:"+id]
	assert.True(t, ok, "expected expirer to be called for key identifier")
}

func TestMiddleware_TooManyRequests_SetsRetryAfterAnd429(t *testing.T) {
	vk := newFakeValkey()
	ks := &fakeKeyStore{limits: map[[32]byte]int{}}
	rl := NewRateLimiter(vk, ks, 2, 45*time.Second, nil)
	mw := rl.Middleware(okHandler())

	// 1st request -> ok, remaining 1
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "198.51.100.1:5678"
	rec1 := httptest.NewRecorder()
	mw.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Result().StatusCode)

	// 2nd -> ok, remaining 0
	rec2 := httptest.NewRecorder()
	mw.ServeHTTP(rec2, req1.Clone(req1.Context()))
	assert.Equal(t, http.StatusOK, rec2.Result().StatusCode)
	assert.Equal(t, "0", rec2.Result().Header.Get("X-RateLimit-Remaining"))

	// 3rd -> 429 and Retry-After header set to window seconds
	rec3 := httptest.NewRecorder()
	mw.ServeHTTP(rec3, req1.Clone(req1.Context()))
	res3 := rec3.Result()
	assert.Equal(t, http.StatusTooManyRequests, res3.StatusCode)
	assert.Equal(t, "45", res3.Header.Get("Retry-After"))
	assert.Equal(t, "0", res3.Header.Get("X-RateLimit-Remaining"))
}

func TestMiddleware_FailOpenOnValkeyError(t *testing.T) {
	vk := newFakeValkey()
	vk.errOnIncr = assert.AnError
	ks := &fakeKeyStore{limits: map[[32]byte]int{}}
	rl := NewRateLimiter(vk, ks, 1, time.Minute, nil)
	mw := rl.Middleware(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:1111"
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	res := rec.Result()
	// Should pass through as OK even though valkey errored
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestRealIP_UntrustedProxy_HeadersIgnored(t *testing.T) {
	vk := newFakeValkey()
	ks := &fakeKeyStore{limits: map[[32]byte]int{}}
	// No trusted proxies configured — forwarded headers must be ignored.
	rl := NewRateLimiter(vk, ks, 10, time.Minute, nil)
	mw := rl.Middleware(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.5:1234"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("X-Real-Ip", "5.6.7.8")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	// Must use RemoteAddr, not the spoofed headers.
	_, spoofed1 := vk.counts["rl:ip:1.2.3.4"]
	_, spoofed2 := vk.counts["rl:ip:5.6.7.8"]
	assert.False(t, spoofed1, "X-Forwarded-For must not be trusted from untrusted RemoteAddr")
	assert.False(t, spoofed2, "X-Real-Ip must not be trusted from untrusted RemoteAddr")
	_, ok := vk.counts["rl:ip:203.0.113.5"]
	assert.True(t, ok, "RemoteAddr should be used as the rate-limit key")
}

func TestRealIP_TrustedProxy_XFFUsed(t *testing.T) {
	vk := newFakeValkey()
	ks := &fakeKeyStore{limits: map[[32]byte]int{}}
	// 10.0.0.0/8 is trusted (internal load balancer subnet).
	rl := NewRateLimiter(vk, ks, 10, time.Minute, []string{"10.0.0.0/8"})
	mw := rl.Middleware(okHandler())

	// X-Forwarded-For first value wins when request arrives from trusted proxy.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.1.1:443"
	req.Header.Set("X-Forwarded-For", "203.0.113.99, 10.0.1.1")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	_, ok := vk.counts["rl:ip:203.0.113.99"]
	assert.True(t, ok, "first XFF value should be used as rate-limit key for trusted proxy")

	// X-Real-Ip is used when XFF is absent.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.1.2:443"
	req2.Header.Set("X-Real-Ip", "198.51.100.7")
	rec2 := httptest.NewRecorder()
	mw.ServeHTTP(rec2, req2)
	_, ok2 := vk.counts["rl:ip:198.51.100.7"]
	assert.True(t, ok2, "X-Real-Ip should be used as rate-limit key for trusted proxy")
}

func TestRealIP_FallbackToRemoteAddr(t *testing.T) {
	vk := newFakeValkey()
	ks := &fakeKeyStore{limits: map[[32]byte]int{}}
	rl := NewRateLimiter(vk, ks, 10, time.Minute, nil)
	mw := rl.Middleware(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.77:9090"
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	_, ok := vk.counts["rl:ip:203.0.113.77"]
	assert.True(t, ok, "RemoteAddr host should be used when no forwarded headers are present")
}
