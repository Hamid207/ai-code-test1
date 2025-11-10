# Docker Setup Guide

Bu layihəni Docker ilə işə salmaq üçün təlimatlar.

## Tələblər

Kompyuterinizdə aşağıdakılar quraşdırılmış olmalıdır:
- [Docker](https://docs.docker.com/get-docker/) (v20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)

## Sürətli Başlanğıc

### 1. Environment faylını hazırlayın

```bash
# .env.example faylını kopyalayın
cp .env.example .env

# .env faylını redaktə edin və lazımi dəyərləri daxil edin
nano .env  # və ya istənilən text editor
```

### 2. Lazımi environment dəyişənlərini konfiqurasiya edin

`.env` faylında aşağıdakıları mütləq dəyişdirin:

```bash
# Güclü parollar təyin edin
DB_PASSWORD=your_strong_password_here
REDIS_PASSWORD=your_redis_password_here

# JWT secret (minimum 32 simvol)
JWT_SECRET=$(openssl rand -base64 48)

# OAuth konfiqurasiyası (Apple/Google Developer Console-dan alın)
APPLE_CLIENT_ID=your_apple_client_id
APPLE_TEAM_ID=your_apple_team_id
GOOGLE_CLIENT_ID=your_google_client_id
```

### 3. Docker konteynerləri işə salın

```bash
# Bütün xidmətləri (PostgreSQL, Redis, App) işə salın
docker-compose up -d

# Logları izləyin
docker-compose logs -f app
```

### 4. Tətbiqin işlədiyini yoxlayın

```bash
# Health check
curl http://localhost:8080/health

# Gözlənilən cavab: {"status":"healthy"}
```

## Docker Əmrləri

### Konteynerləri idarə edin

```bash
# Bütün xidmətləri başlat
docker-compose up -d

# Xidmətləri dayandır
docker-compose stop

# Xidmətləri dayandır və silin
docker-compose down

# Xidmətləri silin və volume-ları da silin (DİQQƏT: Məlumatlar silinəcək!)
docker-compose down -v

# Yenidən build et və başlat
docker-compose up -d --build
```

### Logları görmək

```bash
# Bütün konteynerlərin logları
docker-compose logs

# Yalnız app konteynerinin logları
docker-compose logs app

# Canlı log izləmək
docker-compose logs -f app

# Son 100 sətr
docker-compose logs --tail=100 app
```

### Konteynerin içinə daxil olmaq

```bash
# App konteynerinə daxil ol
docker-compose exec app sh

# PostgreSQL-ə qoşul
docker-compose exec postgres psql -U postgres -d ios_backend

# Redis CLI
docker-compose exec redis redis-cli
```

### Database əməliyyatları

```bash
# PostgreSQL backup
docker-compose exec postgres pg_dump -U postgres ios_backend > backup.sql

# PostgreSQL restore
docker-compose exec -T postgres psql -U postgres ios_backend < backup.sql

# Database-ə manual qoşulmaq
docker-compose exec postgres psql -U postgres -d ios_backend
```

## Xidmətlər

Docker Compose aşağıdakı xidmətləri işə salır:

### 1. **App** (Go Backend)
- Port: `8080`
- Konteyner adı: `ios-backend-app`
- Health check: http://localhost:8080/health

### 2. **PostgreSQL**
- Port: `5432`
- Konteyner adı: `ios-backend-db`
- Database: `ios_backend`
- User: `postgres` (`.env`-dən)

### 3. **Redis**
- Port: `6379`
- Konteyner adı: `ios-backend-redis`
- Max memory: `256mb`
- Persistence: AOF enabled

## Troubleshooting

### Port artıq istifadə olunur

```bash
# İstifadə olunan portu yoxla
lsof -i :8080
lsof -i :5432
lsof -i :6379

# .env faylında portları dəyişdir
SERVER_PORT=8081
DB_PORT=5433
REDIS_PORT=6380
```

### Konteynerlər işə düşmür

```bash
# Konteynerin statusunu yoxla
docker-compose ps

# Detallı logları oxu
docker-compose logs app

# Konteynerləri sıfırdan başlat
docker-compose down
docker-compose up -d --build
```

### Database bağlantı xətası

```bash
# PostgreSQL hazır olduğunu yoxla
docker-compose exec postgres pg_isready -U postgres

# Database mövcudluğunu yoxla
docker-compose exec postgres psql -U postgres -l

# Connection string-i yoxla
docker-compose exec app env | grep DATABASE_URL
```

### Yaddaş problemləri

```bash
# İstifadə olunmayan image-ları təmizlə
docker system prune -a

# Volume-ları təmizlə (DİQQƏT: Məlumatlar silinəcək!)
docker volume prune

# Tam təmizlik
docker system prune -a --volumes
```

## Production Deploy

Production üçün əlavə tövsiyələr:

1. **Güclü parollar istifadə edin**
   ```bash
   openssl rand -base64 32
   ```

2. **SSL/TLS konfiqurasiya edin**
   - Nginx və ya reverse proxy istifadə edin
   - Let's Encrypt sertifikatları

3. **Resource limitləri təyin edin**
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '2'
         memory: 2G
   ```

4. **Secret management**
   - Docker Secrets
   - AWS Secrets Manager
   - HashiCorp Vault

5. **Monitoring və logging**
   - Prometheus + Grafana
   - ELK Stack
   - Sentry

## API Endpoints

Backend işə düşdükdən sonra:

```bash
# Health check
GET http://localhost:8080/health

# Apple OAuth
POST http://localhost:8080/api/v1/auth/apple

# Google OAuth
POST http://localhost:8080/api/v1/auth/google

# Token refresh
POST http://localhost:8080/api/v1/auth/refresh
```

## Kömək

Problemlə qarşılaşsanız:

1. Logları yoxlayın: `docker-compose logs -f`
2. Konteynerin statusunu yoxlayın: `docker-compose ps`
3. Health check-ləri yoxlayın: `curl http://localhost:8080/health`
4. GitHub Issues-da məsələ açın

## Faydalı Linklər

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [PostgreSQL Docker Hub](https://hub.docker.com/_/postgres)
- [Redis Docker Hub](https://hub.docker.com/_/redis)
