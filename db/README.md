# Contful 数据库初始化指南

## 📁 文件结构

```
db/
└── init_pg.sql    # 完整初始化脚本（DDL 建表 + 种子数据）
```

## 🚀 初始化步骤

```bash
# 一条命令完成建表 + 默认数据导入
psql -h <host> -U <user> -d <db> -f db/init_pg.sql
```

**⚠️ 注意**：此脚本会删除所有已有表并重建，**仅用于全新部署或开发环境重置**。

## 🐳 Docker 自动初始化

`entrypoint.sh` 在 `CONTFUL_AUTO_INIT=true` 时自动执行 `init_pg.sql`，无需手动操作。

---

## 📝 init_pg.sql 内容说明

按顺序执行：

1. **启用扩展**：`uuid-ossp`、`pgcrypto`
2. **清理已有对象**：触发器、函数、表、ENUM 类型（支持重复执行）
3. **创建 ENUM 类型**：`user_status`、`entry_status`、`field_type` 等
4. **创建触发器函数**：`update_updated_time_column()`、`prevent_audit_log_update()`
5. **建表**：14 张表（`contful_` 前缀），含字段、索引、触发器、注释
6. **种子数据**：默认站点、系统角色、权限元数据、管理员用户、系统配置（幂等设计）

### 表清单

| 表名 | 说明 |
|---|---|
| `contful_system_users` | 系统用户 |
| `contful_system_roles` | 系统角色 |
| `contful_system_user_roles` | 用户-角色关联 |
| `contful_system_config` | 系统配置 |
| `contful_system_permission_groups` | 权限分组 |
| `contful_system_permissions` | 权限项 |
| `contful_audit_logs` | 审计日志（防篡改） |
| `contful_sites` | 站点 |
| `contful_schemas` | 内容模型 |
| `contful_fields` | 字段定义 |
| `contful_entries` | 内容条目 |
| `contful_entry_values` | 条目字段值 |
| `contful_entry_versions` | 条目版本历史 |
| `contful_asset_folders` | 媒体文件夹 |
| `contful_assets` | 媒体资产 |
| `contful_tokens` | API Token |

---

## 🔑 默认账号

| 字段 | 值 |
|---|---|
| 邮箱 | `admin@contful.com` |
| 密码 | `contful@com` |

> ⚠️ **请登录后立即修改默认密码！**
