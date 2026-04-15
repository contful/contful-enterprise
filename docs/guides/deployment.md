# 部署指南

> 版本: v1.0.0 | 更新日期: 2026-04-15

---

## 1. 部署架构

### 1.1 最小生产架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         公网入口                                 │
│                                                                  │
│              ┌─────────────┐  ┌─────────────┐                    │
│              │   CDN/WAF   │  │   Nginx     │                    │
│              │  (可选)     │  │  (反向代理) │                    │
│              └──────┬──────┘  └──────┬──────┘                    │
│                     │                │                            │
│                     └────────┬───────┘                            │
│                              │                                    │
│              ┌───────────────┴───────────────┐                   │
│              │                               │                    │
│     ┌────────▼────────┐            ┌────────▼────────┐         │
│     │   Admin API     │            │    Open API       │         │
│     │   (:8080)       │            │    (:8081)        │         │
│     │   Console 前端   │            │    第三方接入      │         │
│     └────────┬────────┘            └────────┬────────┘         │
│              │                               │                    │
│              └───────────────┬───────────────┘                   │
│                              │                                    │
│     ┌───────────────┬────────┴────────┬───────────────┐         │
│     │               │                 │               │          │
│     ▼               ▼                 ▼               ▼          │
│ ┌────────┐    ┌────────┐       ┌────────┐    ┌────────┐        │
│ │   PG   │    │  Redis │       │ Local  │    │   OSS  │        │
│ │Primary │    │ Cluster│       │ Storage│    │ (可选)  │        │
│ └────────┘    └────────┘       └────────┘    └────────┘        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. 部署前准备

### 2.1 服务器要求

| 资源 | 最小配置 | 推荐配置 |
|------|----------|----------|
| CPU | 2 核 | 4 核 |
| 内存 | 4 GB | 8 GB |
| 系统盘 | 40 GB | 80 GB SSD |
| 数据盘 | 100 GB | 500 GB SSD |

### 2.2 软件依赖

```bash
# Ubuntu/Debian
apt update && apt install -y \
  docker.io \
  docker-compose \
  nginx \
  certbot

# CentOS/RHEL
yum install -y \
  docker \
  docker-compose \
  nginx \
  certbot
```

---

## 3. Docker Compose 部署

### 3.1 创建部署目录

```bash
mkdir -p /opt/contful
cd /opt/contful
```

### 3.2 创建 docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:18
    container_name: contful-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: contful
      POSTGRES_USER: contful
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U contful"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: contful-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  admin-api:
    image: contful/admin-api:latest
    container_name: contful-admin-api
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_NAME: contful
      DB_USER: contful
      DB_PASSWORD: ${DB_PASSWORD}
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - uploads:/app/uploads

  open-api:
    image: contful/open-api:latest
    container_name: contful-open-api
    restart: unless-stopped
    ports:
      - "127.0.0.1:8081:8081"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_NAME: contful
      DB_USER: contful
      DB_PASSWORD: ${DB_PASSWORD}
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

volumes:
  postgres_data:
  redis_data:
  uploads:
```

### 3.3 创建 .env 文件

```bash
# 数据库密码 (至少 32 位随机字符)
DB_PASSWORD=$(openssl rand -base64 32)

# Redis 密码
REDIS_PASSWORD=$(openssl rand -base64 32)

# JWT 密钥 (至少 64 位随机字符)
JWT_SECRET=$(openssl rand -base64 64)

cat > .env << EOF
DB_PASSWORD=${DB_PASSWORD}
REDIS_PASSWORD=${REDIS_PASSWORD}
JWT_SECRET=${JWT_SECRET}
EOF
```

### 3.4 启动服务

```bash
# 拉取镜像
docker-compose pull

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 检查服务状态
docker-compose ps
```

---

## 4. Nginx 配置

### 4.1 创建 Nginx 配置

```bash
cat > /etc/nginx/sites-available/contful << 'EOF'
upstream admin_api {
    server 127.0.0.1:8080;
}

upstream open_api {
    server 127.0.0.1:8081;
}

server {
    listen 80;
    server_name admin.contful.com;

    # 重定向到 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name admin.contful.com;

    # SSL 配置
    ssl_certificate /etc/letsencrypt/live/contful.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/contful.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers on;

    # Gzip 压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript;

    # Admin API 代理
    location /admin/ {
        proxy_pass http://admin_api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 静态文件
    location /assets/ {
        proxy_pass http://admin_api/assets/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # 健康检查
    location /health {
        proxy_pass http://admin_api/health;
        access_log off;
    }
}

server {
    listen 80;
    server_name api.contful.com;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.contful.com;

    # SSL 配置
    ssl_certificate /etc/letsencrypt/live/contful.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/contful.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;

    # Open API 代理
    location / {
        proxy_pass http://open_api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 速率限制
        limit_req zone=api_limit burst=100 nodelay;
    }
}
EOF
```

### 4.2 启用配置

```bash
# 启用站点
ln -s /etc/nginx/sites-available/contful /etc/nginx/sites-enabled/

# 测试配置
nginx -t

# 重载 Nginx
systemctl reload nginx
```

---

## 5. SSL 证书

### 5.1 获取证书

```bash
# 安装 certbot
apt install -y certbot python3-certbot-nginx

# 获取证书
certbot --nginx -d contful.com -d admin.contful.com -d api.contful.com

# 自动续期测试
certbot renew --dry-run
```

---

## 6. 运维命令

### 6.1 日志管理

```bash
# 查看实时日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f admin-api

# 日志轮转
cat > /etc/logrotate.d/contful << 'EOF'
/opt/contful/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0644 root root
    postrotate
        docker-compose -f /opt/contful/docker-compose.yml kill -s USR1
    endscript
}
EOF
```

### 6.2 备份

```bash
# 数据库备份脚本
cat > /opt/contful/backup.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR=/opt/contful/backups

# PostgreSQL 备份
docker exec contful-postgres pg_dump -U contful contful | gzip > ${BACKUP_DIR}/contful_${DATE}.sql.gz

# 保留最近 30 天
find ${BACKUP_DIR} -name "contful_*.sql.gz" -mtime +30 -delete

echo "Backup completed: ${DATE}"
EOF

chmod +x /opt/contful/backup.sh

# 添加到 crontab
echo "0 2 * * * /opt/contful/backup.sh" | tee -a /var/spool/cron/crontabs/root
```

### 6.3 更新服务

```bash
# 拉取最新镜像
docker-compose pull

# 重启服务
docker-compose up -d --no-deps

# 重建特定服务
docker-compose up -d --build admin-api
```

---

## 7. 监控配置

### 7.1 Prometheus 指标

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'contful'
    static_configs:
      - targets: ['localhost:8080', 'localhost:8081']
    metrics_path: /metrics
```

### 7.2 健康检查

```bash
# 就绪检查
curl http://localhost:8080/health/ready

# 存活检查
curl http://localhost:8080/health/live
```

---

## 8. 故障排查

### 8.1 常见问题

| 问题 | 解决方案 |
|------|----------|
| 服务启动失败 | 检查 `.env` 配置 |
| 数据库连接失败 | 确认 PostgreSQL 是否启动 |
| 502 错误 | 检查 API 服务状态和 Nginx 配置 |
| 上传失败 | 检查 `uploads` 目录权限 |

### 8.2 调试模式

```bash
# 启用调试日志
docker-compose exec admin-api env LOG_LEVEL=debug

# 查看详细错误
docker-compose logs admin-api --tail=100
```
