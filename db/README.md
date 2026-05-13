# Contful 数据库初始化指南

## 📁 文件结构

```
db/
├── init_pg.sql    # 数据库完整初始化脚本（建表 + 触发器 + 索引 + 注释）
├── seed_data.sql  # 初始化种子数据（系统角色、默认管理员、默认站点等）
└── README.md      # 本文档
```

## 🚀 初始化步骤

### 第一步：创建 schema（建表）

```bash
psql -h <host> -U <user> -d <db> -f db/init_pg.sql
```

**⚠️ 注意**：此脚本会删除所有已有表并重建，**仅用于全新部署或开发环境重置**。

### 第二步：写入种子数据

```bash
psql -h <host> -U <user> -d <db> -f db/seed_data.sql
```

`seed_data.sql` 使用幂等设计（`INSERT ... WHERE NOT EXISTS`），可重复执行，不会产生重复数据。

### 一键执行

```bash
psql -h <host> -U <user> -d <db> \
  -f db/init_pg.sql \
  -f db/seed_data.sql
```

### Docker 自动初始化

将两个文件按顺序放到 `docker-entrypoint-initdb.d/`，PostgreSQL 容器启动时会自动按文件名顺序执行：

```dockerfile
COPY db/init_pg.sql  /docker-entrypoint-initdb.d/01-init.sql
COPY db/seed_data.sql /docker-entrypoint-initdb.d/02-seed.sql
```

---

## 📝 init_pg.sql 内容说明

包含以下内容，按顺序执行：

1. **启用扩展**：`uuid-ossp`、`pgcrypto`
2. **清理已有对象**：触发器、函数、表、ENUM 类型（支持重复执行）
3. **创建 ENUM 类型**：`user_status`、`entry_status`、`field_type` 等
4. **创建触发器函数**：`update_updated_time_column()`、`prevent_audit_log_update()`
5. **建表**：14 张表，含字段、索引、触发器、注释
6. **高级索引**：组合索引、部分索引、GIN 全文搜索索引

### 表清单

| 表名 | 说明 |
|---|---|
| `system_users` | 系统用户 |
| `system_roles` | 系统角色 |
| `system_user_roles` | 用户-角色关联 |
| `system_config` | 系统配置 |
| `audit_logs` | 审计日志（防篡改） |
| `sites` | 站点 |
| `schemas` | 内容模型 |
| `fields` | 字段定义 |
| `entries` | 内容条目 |
| `entry_values` | 条目字段值 |
| `entry_versions` | 条目版本历史 |
| `asset_folders` | 媒体文件夹 |
| `assets` | 媒体资产 |
| `tokens` | API Token |

---

## 🔑 默认账号

种子数据中包含一个默认管理员账号：

| 字段 | 值 |
|---|---|
| 邮箱 | `admin@contful.com` |
| 密码 | `Admin@123` |

**请登录后立即修改默认密码！**
