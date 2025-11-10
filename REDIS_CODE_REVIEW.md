# Redis Integration - Professional Code Review

**Review Date**: 2025-11-10
**Reviewer**: Claude Code
**Severity Levels**: 🔴 CRITICAL | 🟠 HIGH | 🟡 MEDIUM | 🟢 LOW

---

## 🔴 CRITICAL ISSUES

### 1. **SCAN Operation in DeleteAllUserTokens - Production Blocker**
**File**: `pkg/redis/token_repository.go:72-101`
**Severity**: 🔴 CRITICAL

```go
func (r *TokenRepository) DeleteAllUserTokens(ctx context.Context, userID int64) error {
    pattern := r.keyBuilder.RefreshTokenPattern(strconv.FormatInt(userID, 10))

    var cursor uint64
    for {
        keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
        // ... işləmə davam edir
        if cursor == 0 {
            break
        }
    }
}
```

**Problemlər**:
1. **SCAN bloklaşdırıcıdır**: Production Redis server-ini yavaşladır
2. **Context timeout check yoxdur**: Sonsuz loop ola bilər
3. **DoS attack vektoru**: Çox token olan user üçün server hang edə bilər
4. **Atomicity problemi**: SCAN zamanı yeni token-lər əlavə oluna bilər

**Impact**: Production Redis performance degradation, potential service outage

**Həll yolu**:
```go
// Option 1: SET data structure istifadə et
// refresh:user:<user_id>:tokens = SET{token_id1, token_id2, ...}

// Option 2: TTL-based cleanup + counter
// Hər token üçün ayrıca TTL, user-level counter

// Option 3: Context deadline check
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // SCAN operation
    }
}
```

---

### 2. **Race Condition in Rate Limiter**
**File**: `pkg/redis/ratelimit_repository.go:52-82`
**Severity**: 🔴 CRITICAL

```go
func (r *RateLimitRepository) incrementRequest(...) {
    // Line 63: Lua script INCR + EXPIRE
    result, err := r.client.Eval(ctx, luaScript, []string{key}, int(window.Seconds())).Result()

    // Line 74: Ayrı Redis call
    ttl, err := r.client.TTL(ctx, key).Result()

    // Line 79: Race condition burada
    resetTime := time.Now().Add(ttl)
}
```

**Problem**:
- Lua script və TTL çağrışı arasında key expire ola bilər
- İki ayrı Redis round-trip inefficient
- `time.Now()` calling time və actual TTL arasında mismatch

**Proof of Concept**:
```
T0: INCR + EXPIRE 1 saniyə (Lua script)
T1: 0.5 saniyə gözləmə (network latency)
T2: TTL əldə et → 0.5 saniyə qalıb
T3: time.Now().Add(0.5) → İNDİKİ ZAMANDAN 0.5 saniyə sonra
Problem: Actual expiry T0 + 1 saniyə, amma biz hesablayırıq T3 + 0.5
```

**Həll yolu**:
```lua
-- Lua script-ə əlavə et
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('TTL', KEYS[1])
return {current, ttl}  -- İki dəyər return et
```

---

## 🟠 HIGH SEVERITY ISSUES

### 3. **String Error Comparison Anti-Pattern**
**File**: `pkg/redis/ratelimit_repository.go:89`
**Severity**: 🟠 HIGH

```go
if err.Error() == "redis: nil" {
    return 0, nil
}
```

**Problem**:
- String comparison fragile və unreliable
- Redis library version update zamanı error message dəyişə bilər
- Performance overhead (string allocation)

**Düzgün yolu**:
```go
if err == redis.Nil {
    return 0, nil
}
```

---

### 4. **Configuration Validation Eksikliyi**
**File**: `pkg/config/config.go:62-76`
**Severity**: 🟠 HIGH

```go
func (c *Config) validate() error {
    // Redis validation YOXDUR!
    if c.AppleClientID == "" {
        return fmt.Errorf("APPLE_CLIENT_ID is required")
    }
    // ...
}
```

**Problem**:
- Redis parametrləri validate olunmur
- İstifadəçi neqativ DB nömrəsi verə bilər
- MaxConns < MinIdleConns ola bilər
- Port validation yoxdur

**Potential issues**:
```bash
REDIS_DB=-1              # Neqativ DB
REDIS_MAX_CONNS=2        # Min-dən kiçik
REDIS_MIN_IDLE_CONNS=10  # Max-dan böyük
REDIS_PORT=abc           # Numeric deyil
```

**Həll yolu**:
```go
func (c *Config) validate() error {
    // Existing validation...

    // Redis validation
    if c.RedisDB < 0 || c.RedisDB > 15 {
        return fmt.Errorf("REDIS_DB must be between 0 and 15")
    }
    if c.RedisMaxConns <= 0 {
        return fmt.Errorf("REDIS_MAX_CONNS must be positive")
    }
    if c.RedisMinIdleConns < 0 {
        return fmt.Errorf("REDIS_MIN_IDLE_CONNS cannot be negative")
    }
    if c.RedisMinIdleConns > c.RedisMaxConns {
        return fmt.Errorf("REDIS_MIN_IDLE_CONNS cannot exceed REDIS_MAX_CONNS")
    }
    // Port validation
    port, err := strconv.Atoi(c.RedisPort)
    if err != nil || port < 1 || port > 65535 {
        return fmt.Errorf("REDIS_PORT must be valid port number (1-65535)")
    }

    return nil
}
```

---

### 5. **Client Config Validation Yoxdur**
**File**: `pkg/redis/client.go:27-60`
**Severity**: 🟠 HIGH

```go
func NewClient(cfg Config) (*Client, error) {
    addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

    rdb := redis.NewClient(&redis.Options{
        // Direct assignment, validation yoxdur
        PoolSize:     cfg.MaxConns,
        MinIdleConns: cfg.MinIdleConns,
    })
}
```

**Problem**:
- Neqativ dəyərlər panic edə bilər
- 0 pool size deadlock yarada bilər
- İstifadəçi `:6379` (boş host) verə bilər

**Həll yolu**:
```go
func NewClient(cfg Config) (*Client, error) {
    // Validate config
    if cfg.Host == "" {
        return nil, fmt.Errorf("redis host cannot be empty")
    }
    if cfg.MaxConns <= 0 {
        return nil, fmt.Errorf("max connections must be positive, got %d", cfg.MaxConns)
    }
    if cfg.MinIdleConns < 0 {
        return nil, fmt.Errorf("min idle connections cannot be negative")
    }
    if cfg.MinIdleConns > cfg.MaxConns {
        return nil, fmt.Errorf("min idle (%d) cannot exceed max (%d)", cfg.MinIdleConns, cfg.MaxConns)
    }
    // ...
}
```

---

## 🟡 MEDIUM SEVERITY ISSUES

### 6. **KeyBuilder Memory Allocation**
**File**: `pkg/redis/*_repository.go` (bütün repository-lər)
**Severity**: 🟡 MEDIUM

```go
func NewTokenRepository(client *Client) *TokenRepository {
    return &TokenRepository{
        client:     client,
        keyBuilder: NewKeyBuilder(),  // Hər instance üçün yeni KeyBuilder
    }
}
```

**Problem**:
- KeyBuilder stateless struct-dır
- Hər repository instance üçün ayrıca KeyBuilder unnecessary
- Memory footprint artır (4 repository × KeyBuilder size)

**Performans təsiri**: Minimal, amma optimization opportunity

**Həll yolu**:
```go
// Global singleton
var keyBuilder = NewKeyBuilder()

func NewTokenRepository(client *Client) *TokenRepository {
    return &TokenRepository{
        client:     client,
        keyBuilder: keyBuilder,  // Shared instance
    }
}
```

---

### 7. **TTL Calculation Precision Loss**
**File**: `pkg/redis/token_repository.go:29`
**Severity**: 🟡 MEDIUM

```go
func (r *TokenRepository) StoreRefreshToken(..., expiresAt time.Time) error {
    key := r.keyBuilder.RefreshToken(...)
    ttl := time.Until(expiresAt)  // Line 29

    if ttl <= 0 {
        return fmt.Errorf("token already expired")
    }

    err := r.client.Set(ctx, key, tokenHash, ttl).Err()  // Line 36
}
```

**Problem**:
- Line 29 və Line 36 arasında nanosecond-level time keçir
- Network latency zamanı token erkən expire ola bilər
- Race condition edge case

**Nümunə**:
```
Line 29: ttl = 1.000000001 saniyə
Network: 1ms latency
Line 36: actual TTL = 0.999 saniyə (intended 1 saniyə)
```

**Həll yolu**:
```go
ttl := time.Until(expiresAt)
if ttl <= 100*time.Millisecond {  // Safety margin
    return fmt.Errorf("token expires too soon (TTL: %v)", ttl)
}
```

---

### 8. **Lua Script Performance**
**File**: `pkg/redis/ratelimit_repository.go:55-61`
**Severity**: 🟡 MEDIUM

```go
luaScript := `
    local current = redis.call('INCR', KEYS[1])
    if current == 1 then
        redis.call('EXPIRE', KEYS[1], ARGV[1])
    end
    return current
`
result, err := r.client.Eval(ctx, luaScript, []string{key}, int(window.Seconds())).Result()
```

**Problem**:
- Hər request-də Lua script Redis-ə göndərilir
- Script compile overhead
- String allocation hər call-da

**Həll yolu**:
```go
// Package level constant
const rateLimitScript = `
    local current = redis.call('INCR', KEYS[1])
    if current == 1 then
        redis.call('EXPIRE', KEYS[1], ARGV[1])
    end
    local ttl = redis.call('TTL', KEYS[1])
    return {current, ttl}
`

// Script SHA cache ilə
var rateLimitScriptSHA string

func init() {
    // Script SHA-sını cache et (SCRIPT LOAD)
}

// İstifadə: EVALSHA əvəzinə EVAL
result, err := r.client.EvalSha(ctx, rateLimitScriptSHA, []string{key}, ...).Result()
```

---

### 9. **Context Propagation Missing**
**File**: `pkg/redis/token_repository.go:72-101`
**Severity**: 🟡 MEDIUM

```go
for {
    keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
    // Context deadline check YOXDUR
    if cursor == 0 {
        break
    }
}
```

**Problem**:
- Context cancel/timeout check edilmir
- User request timeout olsa da loop davam edir
- Resource leak potensialı

**Həll yolu**:
```go
for {
    select {
    case <-ctx.Done():
        return fmt.Errorf("operation cancelled: %w", ctx.Err())
    default:
    }

    keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
    // ...
}
```

---

## 🟢 LOW SEVERITY / CODE QUALITY ISSUES

### 10. **Unused Variable**
**File**: `pkg/redis/token_repository.go:77`
**Severity**: 🟢 LOW

```go
var deletedCount int64

for {
    // ...
    deletedCount += deleted  // Assigned but never used
}

return nil  // deletedCount return olunmur
```

**Həll**: Ya return et, ya da silib `_` istifadə et

---

### 11. **Error Wrapping Inconsistency**
**File**: Multiple files
**Severity**: 🟢 LOW

```go
// Bəzi yerlərdə
return fmt.Errorf("failed to X: %w", err)

// Digər yerlərdə
return fmt.Errorf("failed to X: %v", err)  // Stack trace itir
```

**Best practice**: Həmişə `%w` istifadə et

---

### 12. **Magic Numbers**
**File**: `pkg/redis/token_repository.go:83`
**Severity**: 🟢 LOW

```go
keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
//                                                        ^^^ Magic number
```

**Həll**:
```go
const scanBatchSize = 100

keys, cursor, err = r.client.Scan(ctx, cursor, pattern, scanBatchSize).Result()
```

---

### 13. **Missing Logging**
**File**: Bütün repository-lər
**Severity**: 🟢 LOW

**Problem**: Heç bir Redis operation log edilmir

**Təklif**:
```go
func (r *TokenRepository) StoreRefreshToken(...) error {
    logger.Debug("storing refresh token",
        zap.Int64("user_id", userID),
        zap.String("token_id", tokenID),
        zap.Duration("ttl", ttl),
    )
    // ...
}
```

---

## 📊 SUMMARY

| Severity | Count | Issues |
|----------|-------|---------|
| 🔴 CRITICAL | 2 | SCAN blocking, Race condition |
| 🟠 HIGH | 3 | String error compare, Config validation ×2 |
| 🟡 MEDIUM | 4 | Memory allocation, TTL precision, Lua perf, Context |
| 🟢 LOW | 4 | Unused var, Error wrapping, Magic numbers, Logging |
| **TOTAL** | **13** | |

---

## 🎯 RECOMMENDED ACTIONS

### Immediate (Before Production):
1. ✅ Fix SCAN operation - Use SET data structure
2. ✅ Fix race condition in rate limiter
3. ✅ Add config validation for Redis params
4. ✅ Fix string error comparison

### Short-term (Next Sprint):
5. Add context deadline checks
6. Optimize Lua script with EVALSHA
7. Add structured logging
8. Fix TTL precision with safety margin

### Long-term (Technical Debt):
9. KeyBuilder singleton pattern
10. Metrics/monitoring integration
11. Integration tests with Redis
12. Performance benchmarks

---

## 🧪 TESTING RECOMMENDATIONS

```go
// Test case-lər əlavə et:

func TestDeleteAllUserTokens_ContextTimeout(t *testing.T) {
    // Context timeout zamanı graceful exit
}

func TestIncrementRequest_RaceCondition(t *testing.T) {
    // Concurrent requests zamanı accurate count
}

func TestStoreRefreshToken_AlreadyExpired(t *testing.T) {
    // Expired token reject olunmalı
}

func TestRateLimiter_ResetTimeAccuracy(t *testing.T) {
    // Reset time accurate olmalı
}
```

---

## 📈 PERFORMANCE METRICS (Predicted)

| Operation | Current | After Fix | Improvement |
|-----------|---------|-----------|-------------|
| DeleteAllUserTokens (1000 tokens) | ~2-5s | ~50ms | 40-100x |
| IncrementRequest | 2 RTT | 1 RTT | 2x |
| Rate limiter accuracy | ±100ms | ±10ms | 10x |

---

## 🔐 SECURITY CONSIDERATIONS

✅ **Good**:
- Token hash storage (not plaintext)
- TTL-based automatic cleanup
- Context timeout support (partially)

⚠️ **Concerns**:
- No rate limiting on SCAN operations
- No audit logging
- Missing input sanitization for IP addresses

---

## 📝 CODE REVIEW CONCLUSION

**Overall Assessment**: 🟡 **GOOD with CRITICAL FIXES NEEDED**

Kod Clean Architecture-ə uyğundur və strukturu peşəkardır, **AMMA**:
- 2 critical issue production-da problem yaradacaq
- Config validation missing
- Performance optimization lazımdır

**Recommendation**: ✋ **DO NOT MERGE** until critical issues are fixed.

---

**Next Steps**:
1. Fix critical issues (Priority 1)
2. Add integration tests
3. Performance benchmarks
4. Re-review after fixes

**Estimated Fix Time**: 4-6 hours

