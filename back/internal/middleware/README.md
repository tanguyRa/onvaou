# middleware

```tree
middleware/
├── README.md
├── chain.go
│   ├── type Constructor func(http.Handler) http.Handler
│   ├── type Chain {constructors: []Constructor}
│   ├── func New(constructors ...Constructor) Chain
│   ├── func (Chain) Then(h http.Handler) http.Handler
│   ├── func (Chain) ThenFunc(fn http.HandlerFunc) http.Handler
│   ├── func (Chain) Append(constructors ...Constructor) Chain
│   └── func (Chain) Extend(chain Chain) Chain
├── chain_test.go
│   ├── func tagMiddleware(tag string) Constructor
│   ├── func funcsEqual(f1, f2 interface{}) bool
│   ├── func TestNew(t *testing.T)
│   ├── func TestThenWorksWithNoMiddleware(t *testing.T)
│   ├── func TestThenTreatsNilAsDefaultServeMux(t *testing.T)
│   ├── func TestThenFuncTreatsNilAsDefaultServeMux(t *testing.T)
│   ├── func TestThenFuncConstructsHandlerFunc(t *testing.T)
│   ├── func TestThenOrdersHandlersCorrectly(t *testing.T)
│   ├── func TestAppendAddsHandlersCorrectly(t *testing.T)
│   ├── func TestAppendRespectsImmutability(t *testing.T)
│   ├── func TestExtendAddsHandlersCorrectly(t *testing.T)
│   └── func TestExtendRespectsImmutability(t *testing.T)
├── cors.go
│   ├── func allowedOrigins() []string
│   ├── func isOriginAllowed(origin string, allowed []string) bool
│   └── func CORS(next http.Handler) http.Handler
├── ratelimit.go
│   ├── type RateLimiter {mu: sync.RWMutex, buckets: map[string]*tokenBucket, rate: int, interval: time.Duration, burst: int, stopCh: chan struct{}}
│   ├── type tokenBucket {tokens: int, lastRefill: time.Time}
│   ├── type UserKeyExtractor func(r *http.Request) string
│   ├── func NewRateLimiter(rate int, interval time.Duration, burst int) *RateLimiter
│   ├── func (*RateLimiter) Stop()
│   ├── func (*RateLimiter) Allow(key string) bool
│   ├── func (*RateLimiter) cleanup()
│   ├── func RateLimitMiddleware(rl *RateLimiter, keyExtractor UserKeyExtractor) Constructor
│   └── func getClientIP(r *http.Request) string
└── ratelimit_test.go
    ├── func TestRateLimiter_Allow(t *testing.T)
    ├── func TestRateLimiter_Refill(t *testing.T)
    ├── func TestRateLimiter_DifferentKeys(t *testing.T)
    ├── func TestRateLimitMiddleware(t *testing.T)
    ├── func TestRateLimitMiddleware_FallbackToIP(t *testing.T)
    ├── func TestRateLimitMiddleware_ResponseHeaders(t *testing.T)
    ├── func TestGetClientIP(t *testing.T)
    ├── func TestRateLimiter_Stop(t *testing.T)
    └── func BenchmarkRateLimiter_Allow(b *testing.B)
```
