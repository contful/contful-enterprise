# Contful Enterprise 数据库初始化指南

## 📁 文件结构

```
db/
├── init_pg.sql              # 社区版 PostgreSQL 初始化
├── init_ent_pg.sql           # 企业版 PostgreSQL 增量（审计导出、保留策略）
├── init_ent_dm.sql           # 企业版达梦 DM8 完整初始化
├── upgrade/                  # 升级脚本目录
│   ├── upgrade_v1.4.0_webhook_pg.sql   # v1.4.0 Webhook PG 升级
│   └── upgrade_v1.4.0_webhook_dm.sql   # v1.4.0 Webhook DM 升级
└── README.md
```

## 🚀 PostgreSQL 初始化

```bash
# 企业版部署：先社区版基础表，再企业版增量
psql -h <host> -U <user> -d contful_ent -f db/init_pg.sql
psql -h <host> -U <user> -d contful_ent -f db/init_ent_pg.sql
```

## 🚀 达梦 DM8 初始化

```sql
-- 以 SYSDBA 身份执行一条 SQL 文件即可
-- dm://SYSDBA:SYSDBA008@139.198.171.102:5236
DROP USER CONTFUL_ENT CASCADE;
@db/init_ent_dm.sql
```

## 🐳 Docker 自动初始化

设置 `CONTFUL_AUTO_INIT=true` + `DB_TYPE=dm`，服务启动时自动选择对应 SQL：

| DB_TYPE | 执行的 SQL |
|---------|-----------|
| `postgres`（默认） | `init_pg.sql` + `init_ent_pg.sql` |
| `dm` | `init_ent_dm.sql` |

## 🔑 默认账号

| 字段 | 值 |
|------|-----|
| 邮箱 | `admin@contful.com` |
| 密码 | `contful@com` |

> ⚠️ 登录后立即修改默认密码
