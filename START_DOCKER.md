# 🚀 Docker Konteynerləri İşə Salma Təlimatı

## İndi edin:

```bash
# 1. Köhnə konteynerləri və volume-ları tamamilə silin (təmiz başlanğıc)
docker-compose down -v

# 2. Konteynerləri yenidən başladın
docker-compose up -d

# 3. Konteynerlər işə düşənədək gözləyin (20-30 saniyə)
sleep 30

# 4. Status yoxlayın
docker-compose ps

# 5. Logları izləyin
docker-compose logs -f app
```

## Gözlənilən nəticə:

```
NAME                 IMAGE                    STATUS         PORTS
ios-backend-app      ai-code-test1-app        Up (healthy)   0.0.0.0:8080->8080/tcp
ios-backend-db       postgres:16-alpine       Up (healthy)   0.0.0.0:5432->5432/tcp
ios-backend-redis    redis:7-alpine           Up (healthy)   0.0.0.0:6379->6379/tcp
```

## Test edin:

```bash
# Health check
curl http://localhost:8080/health

# Gözlənilən cavab:
{"status":"healthy"}
```

## Əgər problem olarsa:

```bash
# Redis loglarını oxuyun
docker-compose logs redis

# PostgreSQL loglarını oxuyun
docker-compose logs postgres

# App loglarını oxuyun
docker-compose logs app

# Konteynerin içinə daxil olun
docker-compose exec redis sh
docker-compose exec postgres sh
docker-compose exec app sh
```

## Debug Redis:

```bash
# Redis-ə qoşulun (parol tələb olunur)
docker-compose exec redis redis-cli

# Redis içində:
AUTH 2aW8eR1tY4uI7oP0sAf8K2mP9nQ4rT7w
PING
# Cavab: PONG
```

## Debug PostgreSQL:

```bash
# PostgreSQL-ə qoşulun
docker-compose exec postgres psql -U postgres -d ios_backend

# PostgreSQL içində:
\l          # List databases
\dt         # List tables
\q          # Quit
```
