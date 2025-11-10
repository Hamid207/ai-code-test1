# Redis Integration

Bu paket Redis inteqrasiyasını Clean Architecture prinsiplərinə uyğun olaraq həyata keçirir.

## Arxitektura

### Layer Strukturu

```
pkg/redis/                  # Infrastructure Layer
├── client.go              # Redis connection client
├── keys.go                # Key naming strategy
├── token_repository.go    # Refresh token storage implementation
├── blacklist_repository.go # Token blacklist implementation
├── ratelimit_repository.go # Rate limiting implementation
└── cache_repository.go    # Caching implementation

internal/repository/
└── redis_repository.go    # Repository interfaces (Domain Layer)
```

## Xüsusiyyətlər

### 1. Refresh Token İdarəetməsi

**Məqsəd**: Refresh token-ləri TTL (Time To Live) ilə saxlamaq

**Açar Formatı**: `refresh:<user_id>:<token_id>`

**Interface**: `RedisTokenRepository`

**Metodlar**:
- `StoreRefreshToken()` - Refresh token saxlamaq (avtomatik TTL ilə)
- `GetRefreshToken()` - Refresh token əldə etmək
- `DeleteRefreshToken()` - Bir token silmək
- `DeleteAllUserTokens()` - İstifadəçinin bütün token-lərini silmək

**İstifadə nümunəsi**:
```go
// Refresh token saxla (7 gün TTL)
expiresAt := time.Now().Add(7 * 24 * time.Hour)
err := redisTokenRepo.StoreRefreshToken(ctx, userID, tokenID, tokenHash, expiresAt)

// Token əldə et
tokenHash, err := redisTokenRepo.GetRefreshToken(ctx, userID, tokenID)

// Token sil (logout zamanı)
err := redisTokenRepo.DeleteRefreshToken(ctx, userID, tokenID)
```

### 2. Token Blacklist (Qara Siyahı)

**Məqsəd**: Logout zamanı və ya revoke edilmiş JWT token-ləri qeyd etmək

**Açar Formatı**: `blacklist:<token_id>`

**Interface**: `RedisBlacklistRepository`

**Metodlar**:
- `AddToBlacklist()` - Token-i blacklist-ə əlavə etmək (TTL ilə)
- `IsBlacklisted()` - Token-in blacklist-də olub-olmadığını yoxlamaq

**İstifadə nümunəsi**:
```go
// Logout zamanı access token-i blacklist-ə əlavə et
err := redisBlacklistRepo.AddToBlacklist(ctx, tokenID, expiresAt)

// Middleware-də token-in blacklist-də olub-olmadığını yoxla
isBlacklisted, err := redisBlacklistRepo.IsBlacklisted(ctx, tokenID)
if isBlacklisted {
    return errors.New("token is blacklisted")
}
```

### 3. Rate Limiting

**Məqsəd**: İstifadəçi və ya IP əsaslı sorğu limitləri

**Açar Formatları**:
- `ratelimit:user:<user_id>` - İstifadəçi əsaslı
- `ratelimit:ip:<ip_address>` - IP əsaslı

**Interface**: `RedisRateLimitRepository`

**Metodlar**:
- `IncrementUserRequest()` - İstifadəçi sorğu sayını artırmaq
- `IncrementIPRequest()` - IP sorğu sayını artırmaq
- `GetUserRequestCount()` - Cari istifadəçi sorğu sayı
- `GetIPRequestCount()` - Cari IP sorğu sayı

**İstifadə nümunəsi**:
```go
// İstifadəçi üçün rate limiting (1 dəqiqə window)
window := 1 * time.Minute
count, resetTime, err := redisRateLimitRepo.IncrementUserRequest(ctx, userID, window)
if count > 10 {
    return errors.New("rate limit exceeded")
}

// IP üçün rate limiting
count, resetTime, err := redisRateLimitRepo.IncrementIPRequest(ctx, ipAddress, window)
```

**Xüsusiyyətlər**:
- Atom operasiyalar (Lua script istifadəsi)
- Avtomatik TTL idarəetməsi
- Window-based counting
- Reset time məlumatı

### 4. Caching

**Məqsəd**: İstifadəçi profili və statik məlumatların keşləşdirilməsi

**Açar Formatları**:
- `cache:user:<user_id>` - İstifadəçi məlumatları
- `cache:profile:<user_id>` - Profil məlumatları
- Custom açarlar üçün generic metodlar

**Interface**: `RedisCacheRepository`

**Metodlar**:
- `SetUserCache()` - İstifadəçi məlumatını keşləmək
- `GetUserCache()` - İstifadəçi məlumatını keşdən əldə etmək
- `InvalidateUserCache()` - Keşi etibarsız etmək
- `SetGeneric()` - İstənilən JSON məlumatı keşləmək
- `GetGeneric()` - JSON məlumatı deserialize edərək əldə etmək
- `Delete()` - Keş açarını silmək

**İstifadə nümunəsi**:
```go
// İstifadəçi məlumatını keşlə (5 dəqiqə TTL)
err := redisCacheRepo.SetUserCache(ctx, userID, user, 5*time.Minute)

// Keşdən oxu
user, err := redisCacheRepo.GetUserCache(ctx, userID)

// Cache invalidation (user update zamanı)
err := redisCacheRepo.InvalidateUserCache(ctx, userID)

// Generic keşləmə
err := redisCacheRepo.SetGeneric(ctx, "custom:key", data, 10*time.Minute)
```

## Konfiqurasiya

### Environment Variables

```bash
REDIS_HOST=localhost          # Redis host (default: localhost)
REDIS_PORT=6379              # Redis port (default: 6379)
REDIS_DB=0                   # Redis database number (default: 0)
REDIS_PASSWORD=              # Redis password (optional)
REDIS_MAX_CONNS=10           # Maximum connections (default: 10)
REDIS_MIN_IDLE_CONNS=2       # Minimum idle connections (default: 2)
```

### Connection Pool Parametrləri

- **DialTimeout**: 5 saniyə
- **ReadTimeout**: 3 saniyə
- **WriteTimeout**: 3 saniyə
- **PoolTimeout**: 4 saniyə
- **ConnMaxIdleTime**: 5 dəqiqə
- **MaxRetries**: 3
- **MinRetryBackoff**: 8ms
- **MaxRetryBackoff**: 512ms

## Açar Adlandırma Strategiyası

Bütün Redis açarları aşağıdakı formatda strukturlaşdırılıb:

```
<namespace>:<entity>:<identifier>
```

### Namespace-lər

1. **refresh** - Refresh token-lər
   - Format: `refresh:<user_id>:<token_id>`
   - Nümunə: `refresh:12345:a8f7b3c9-1234-5678-90ab-cdef12345678`

2. **blacklist** - Qara siyahıda olan token-lər
   - Format: `blacklist:<token_id>`
   - Nümunə: `blacklist:a8f7b3c9-1234-5678-90ab-cdef12345678`

3. **token_family** - Token family tracking
   - Format: `token_family:<family_id>`
   - Nümunə: `token_family:f1a2b3c4-5678-90ab-cdef-123456789012`

4. **ratelimit:user** - İstifadəçi rate limiting
   - Format: `ratelimit:user:<user_id>`
   - Nümunə: `ratelimit:user:12345`

5. **ratelimit:ip** - IP rate limiting
   - Format: `ratelimit:ip:<ip_address>`
   - Nümunə: `ratelimit:ip:192.168.1.100`

6. **cache:user** - İstifadəçi keşi
   - Format: `cache:user:<user_id>`
   - Nümunə: `cache:user:12345`

7. **cache:profile** - Profil keşi
   - Format: `cache:profile:<user_id>`
   - Nümunə: `cache:profile:12345`

## Clean Architecture Alignment

### Dependencies

```
Domain Layer (internal/repository/redis_repository.go)
  ↑ Interfaces defined
  |
  ↓ Implemented by
Infrastructure Layer (pkg/redis/*)
```

### Dependency Rule

- **Infrastructure Layer** (pkg/redis) domain layer-dən asılıdır
- **Domain Layer** (internal/repository) heç bir external framework-dən asılı deyil
- Interfaces domain layer-də, implementation infrastructure layer-də

## Health Check

Redis connection-un sağlamlığını yoxlamaq:

```go
err := redisClient.HealthCheck(ctx)
if err != nil {
    log.Printf("Redis health check failed: %v", err)
}
```

## Graceful Shutdown

```go
// Server shutdown zamanı Redis connection-u bağla
if err := redisClient.Close(); err != nil {
    log.Printf("Error closing Redis connection: %v", err)
}
```

## TODO: Gələcək İnteqrasiyalar

1. **Auth Service Integration**
   - Refresh token-ləri PostgreSQL əvəzinə Redis-də saxlamaq
   - Token rotation zamanı köhnə token-ləri Redis-dən silmək

2. **JWT Middleware**
   - Access token-lərin blacklist-də olub-olmadığını yoxlamaq
   - Logout zamanı token-i blacklist-ə əlavə etmək

3. **Logout Handler**
   - Access token-i blacklist-ə əlavə etmək
   - Refresh token-ləri Redis-dən silmək

4. **Rate Limiting Middleware**
   - In-memory limiter əvəzinə Redis-based limiter istifadə etmək
   - Distributed rate limiting dəstəyi

5. **User Profile Caching**
   - Tez-tez oxunan user məlumatlarını keşləmək
   - Cache invalidation strategiyası

## Testing

Redis inteqrasiyasını test etmək üçün:

```bash
# Redis server-i local olaraq işə sal
docker run -d -p 6379:6379 redis:latest

# Application-ı build et
go build -o server ./cmd/server/

# Server-i işə sal
./server
```

Redis connection log-da görünməlidir:
```
Redis connection established successfully (Host: localhost:6379, DB: 0)
```

## Performans Optimizasiyaları

1. **Connection Pooling**: Min 2, Max 10 connection
2. **Automatic Retry**: 3 retry attempt ilə
3. **TTL-based Expiration**: Avtomatik key expiration
4. **Lua Scripts**: Atom operasiyalar üçün
5. **Pipeline Support**: Toplu əməliyyatlar üçün hazır

## Təhlükəsizlik

1. **Password Authentication**: REDIS_PASSWORD dəstəyi
2. **Timeout Configuration**: DoS hücumlarının qarşısını almaq üçün
3. **Connection Limits**: Resource exhaustion-un qarşısını almaq
4. **TTL Enforcement**: Köhnə məlumatların avtomatik silinməsi
