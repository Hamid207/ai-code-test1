# 🔐 TƏHLÜKƏSİZLİK VƏ CREDENTIALS İDARƏETMƏ

## 🚨 KRİTİK XƏBƏRDARLIK

**HƏRGIZ bu faylları commit ETMƏYİN:**
- ❌ `.env`
- ❌ `.env.local`
- ❌ `AuthKey_*.p8` (Apple private keys)
- ❌ Hər hansı API keys, passwords, secrets

---

## ✅ DÜZGÜN İSTİFADƏ

### Addım 1: .env Faylı Yaradın

**Avtomatik (Tövsiyə olunur):**
```bash
# Setup script ilə avtomatik yaradın
./setup-env.sh

# Bu script:
# ✅ Güclü parollar generate edəcək
# ✅ .env faylı yaradacaq
# ✅ Təhlükəsiz konfiqurasiya edəcək
```

**Manual:**
```bash
# .env.example-dan kopyalayın
cp .env.example .env

# Güclü parollar generate edin
DB_PASSWORD=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 48)

# .env faylını redaktə edin
nano .env
```

---

### Addım 2: Real Credentials Əlavə Edin

`.env` faylını açın və bu dəyərləri əlavə edin:

```bash
# Apple Developer Account-dan götürün
APPLE_TEAM_ID=77QNKT8P7A
APPLE_CLIENT_ID=com.hamidmanafov.Micnoteai
APPLE_KEY_ID=YVJ9V9735T
APPLE_PRIVATE_KEY_PATH=./configs/AuthKey_YVJ9V9735T.p8

# Google Cloud Console-dan götürün
GOOGLE_CLIENT_ID=800505339834-6o90ggiulnulu7ejm5k9dj6lbpuj66mg.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=YOUR_GOOGLE_CLIENT_SECRET_HERE

# OpenAI Platform-dan götürün (KÖHNƏ KEY-İ SİLİN!)
OPENAI_API_KEY=sk-proj-NEW_KEY_HERE_AFTER_ROTATING
```

---

### Addım 3: Apple Private Key Yerləşdirin

```bash
# configs qovluğu yaradın
mkdir -p configs

# Apple private key-i yerləşdirin
cp /path/to/AuthKey_YVJ9V9735T.p8 configs/

# İcazələri təyin edin
chmod 600 configs/AuthKey_YVJ9V9735T.p8
```

---

## 🔴 EXPOSE OLMUŞ OPENAI API KEY

**Göndərdiyiniz OpenAI API key artıq public oldu!**

### Dərhal edin:

1. **OpenAI Platform-a gedin:**
   ```
   https://platform.openai.com/api-keys
   ```

2. **Köhnə key-i silin:**
   - Köhnə key-i tapın və "Revoke" düyməsini basın
   - Key: `sk-proj-KxeMe28ty...` (silinməlidir!)

3. **Yeni key generate edin:**
   - "Create new secret key" düyməsini basın
   - Ad verin: "iOS Backend Production"
   - Kopyalayın və `.env` faylına əlavə edin

4. **Billing yoxlayın:**
   ```
   https://platform.openai.com/usage
   ```
   - Unexpected usage yoxlayın
   - Billing alerts təyin edin

---

## 📋 .gitignore Yoxlayın

Əmin olun ki, bu fayllar `.gitignore`-da var:

```bash
# Check .gitignore
cat .gitignore | grep -E "\.env|\.p8|AuthKey"

# Olmalıdır:
.env
.env.local
.env.*.local
*.p8
configs/*.p8
AuthKey_*.p8
```

---

## ✅ Təhlükəsizlik Checklist

Başlamazdan əvvəl yoxlayın:

- [ ] `.env` faylı yaradıldı və `.gitignore`-da var
- [ ] Güclü parollar generate edildi (minimum 32 char)
- [ ] Apple credentials əlavə edildi
- [ ] Google credentials əlavə edildi
- [ ] OpenAI API key **rotate edildi** (köhnəsi silindi!)
- [ ] Apple private key `configs/` qovluğunda və `chmod 600`
- [ ] `.env` faylı **heç vaxt commit edilməyəcək**
- [ ] `git status` - .env faylı "Untracked" olaraq görünmür

---

## 🚀 İstifadə

### Local Development:

```bash
# 1. .env faylı yaradın (yuxarıda göstərildiyi kimi)
./setup-env.sh

# 2. Credentials əlavə edin
nano .env

# 3. Docker containers başladın
docker-compose up -d

# 4. Logları izləyin
docker-compose logs -f app

# 5. Test edin
curl http://localhost:8080/health
```

### Production Deployment:

```bash
# Production-da .env faylı istifadə ETMƏYİN!
# Bunun yerinə:

# AWS: Use AWS Secrets Manager
# Azure: Use Azure Key Vault
# GCP: Use Google Secret Manager
# Docker: Use Docker Secrets
# Kubernetes: Use Kubernetes Secrets
```

---

## 🔒 Production Secrets Management

### AWS Secrets Manager (Tövsiyə):

```bash
# Install AWS CLI
aws configure

# Store secrets
aws secretsmanager create-secret \
    --name ios-backend/db-password \
    --secret-string "$DB_PASSWORD"

aws secretsmanager create-secret \
    --name ios-backend/openai-key \
    --secret-string "$OPENAI_API_KEY"

# Retrieve in app
aws secretsmanager get-secret-value \
    --secret-id ios-backend/db-password \
    --query SecretString --output text
```

### Docker Secrets:

```bash
# Create secrets
echo "$DB_PASSWORD" | docker secret create db_password -
echo "$REDIS_PASSWORD" | docker secret create redis_password -
echo "$JWT_SECRET" | docker secret create jwt_secret -

# Use in docker-compose.yml
services:
  app:
    secrets:
      - db_password
      - jwt_secret

secrets:
  db_password:
    external: true
  jwt_secret:
    external: true
```

---

## 🛡️ Best Practices

### 1. Password Strength:
```bash
# GOOD:
openssl rand -base64 32
# Output: 8K2mP9nQ4rT7wX1zC5vB8nM3kL6jH9dF

# BAD:
postgres
password123
mypassword
```

### 2. API Key Rotation:
```bash
# Hər 90 gündə bir rotate edin
# Calendar reminder təyin edin
```

### 3. Environment Separation:
```bash
.env.development    # Local dev
.env.staging        # Test environment
.env.production     # Production (use secrets manager!)
```

### 4. Audit:
```bash
# Regular audit
git log --all --full-history -- "*.env*"
git log --all --full-history -- "*.p8"

# Təmin edin ki, heç vaxt commit edilməyib
```

---

## 🆘 TƏCILI: Credentials Expose Olubsa

### 1. Dərhal:
- [ ] OpenAI API key rotate edin
- [ ] Google OAuth credentials rotate edin
- [ ] Database parolunu dəyişin
- [ ] Redis parolunu dəyişin
- [ ] JWT secret dəyişin

### 2. Yoxlayın:
- [ ] Billing/usage unexpected activity yoxlayın
- [ ] Access logs yoxlayın
- [ ] Security alerts yoxlayın

### 3. Git Tarixindən silin (Təcili halda):
```bash
# BFG Repo-Cleaner
bfg --delete-files .env
git reflog expire --expire=now --all
git gc --prune=now --aggressive

# Və ya
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch .env" \
  --prune-empty --tag-name-filter cat -- --all
```

---

## 📞 Kömək

Problem olarsa:
1. Bu guide-ı yenidən oxuyun
2. `.gitignore`-ı yoxlayın: `cat .gitignore | grep .env`
3. Git status yoxlayın: `git status`
4. `.env` faylı görünürsə: **COMMIT ETMƏYİN!**

---

## ✅ Summary

| Etməli | Etməməli |
|--------|----------|
| ✅ `.env` faylı yaradın (local) | ❌ `.env` faylını commit edin |
| ✅ Güclü parollar generate edin | ❌ Weak passwords istifadə edin |
| ✅ OpenAI key rotate edin | ❌ Expose olmuş key-ləri istifadə edin |
| ✅ `.gitignore`-da olduğunu yoxlayın | ❌ API keys-ləri hard-code edin |
| ✅ Production-da secrets manager | ❌ Production-da .env faylı |
| ✅ Regular audit/rotation | ❌ Credentials-ları share edin |

---

**Ən önemli qaydalar:**

1. 🔴 **HƏRGIZ `.env` faylını commit etməyin!**
2. 🔴 **Expose olmuş OpenAI key-i dərhal rotate edin!**
3. 🔴 **Production-da secrets manager istifadə edin!**
4. ✅ **Güclü parollar generate edin!**
5. ✅ **Regular rotation təyin edin!**
