# Contful 数据库多数据库兼容性说明

本文档说明 `init_pg.sql` 迁移到 MySQL / 达梦（DM）时需要的改动。

> **注意**：当前项目仅官方支持 PostgreSQL 14+。其他数据库的支持为社区贡献方向，不保证功能完整。

## 一、类型映射

| PostgreSQL | MySQL | 达梦 (DM8) | 备注 |
|-----------|-------|-----------|------|
| `UUID` | `VARCHAR(36)` | `VARCHAR(36)` | MySQL 8.0+ 可用自定义类型 |
| `JSONB` | `JSON` | `CLOB` | MySQL 5.7+ 支持 JSON |
| `TIMESTAMPTZ` | `DATETIME` / `TIMESTAMP` | `TIMESTAMP` | MySQL TIMESTAMP 有 2038 问题 |
| `INET` | `VARCHAR(45)` | `VARCHAR(45)` | 存 IPv4/IPv6 地址 |
| `BOOLEAN` | `TINYINT(1)` | `NUMBER(1)` | MySQL 无原生 BOOLEAN |
| `TEXT` | `TEXT` / `LONGTEXT` | `CLOB` | MySQL TEXT 有 64KB 限制 |
| `user_status` (ENUM) | `VARCHAR(20) CHECK (...)` | `VARCHAR(20) CHECK (...)` | 见下方 ENUM 替换方案 |
| `gen_random_uuid()` | `UUID()` | `SYS_GUID()` | 主键默认值 |
| `NOW()` | `NOW()` / `CURRENT_TIMESTAMP` | `CURRENT_TIMESTAMP` | 通用 |

## 二、ENUM 类型替换

PostgreSQL 使用 `CREATE TYPE ... AS ENUM (...)` 定义枚举类型，MySQL/达梦不支持此语法。

### 替换方案

```sql
-- PostgreSQL
status user_status NOT NULL DEFAULT 'active'

-- MySQL / 达梦
status VARCHAR(20) NOT NULL DEFAULT 'active' 
    CHECK (status IN ('active', 'inactive', 'suspended'))
```

需要替换的 ENUM 类型：
- `user_status` → `VARCHAR(20) + CHECK`
- `entry_status` → `VARCHAR(20) + CHECK`
- `content_type_kind` → `VARCHAR(20) + CHECK`
- `field_type` → `VARCHAR(20) + CHECK`
- `asset_type` → `VARCHAR(20) + CHECK`
- `asset_visibility` → `VARCHAR(20) + CHECK`
- `asset_status` → `VARCHAR(20) + CHECK`
- `token_status` → `VARCHAR(20) + CHECK`
- `audit_level` → `VARCHAR(20) + CHECK`
- `audit_type` → `VARCHAR(20) + CHECK`

## 三、部分索引

PostgreSQL 支持 `WHERE` 条件的部分索引（Partial Index），MySQL 不支持。

```sql
-- PostgreSQL
CREATE UNIQUE INDEX idx_system_users_email_active ON system_users(email) WHERE deleted_time IS NULL;

-- MySQL 替代方案：使用生成的列 + 普通唯一索引
ALTER TABLE system_users ADD COLUMN email_active VARCHAR(255) 
    GENERATED ALWAYS AS (IF(deleted_time IS NULL, email, NULL)) STORED;
CREATE UNIQUE INDEX idx_system_users_email_active ON system_users(email_active);
```

## 四、触发器

PostgreSQL 使用 `plpgsql` 函数 + `CREATE TRIGGER`，MySQL 语法不同。

```sql
-- PostgreSQL: 自定义函数 + 触发器
CREATE OR REPLACE FUNCTION update_updated_time_column() ...

-- MySQL 替代方案 1: ON UPDATE CURRENT_TIMESTAMP（仅适用于单列）
updated_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

-- MySQL 替代方案 2: 触发器
CREATE TRIGGER update_system_users_updated_time 
    BEFORE UPDATE ON system_users 
    FOR EACH ROW SET NEW.updated_time = CURRENT_TIMESTAMP;
```

## 五、ON CONFLICT（Upsert）

```sql
-- PostgreSQL
INSERT INTO system_config (...) VALUES (...) ON CONFLICT (config_key) DO NOTHING;

-- MySQL
INSERT INTO system_config (...) VALUES (...) ON DUPLICATE KEY UPDATE config_key = config_key;

-- 达梦
MERGE INTO system_config t USING (SELECT ... FROM dual) s ON (t.config_key = s.config_key)
WHEN NOT MATCHED THEN INSERT (...) VALUES (...);
```

## 六、GIN 全文搜索索引

```sql
-- PostgreSQL
CREATE INDEX idx_entry_values_text_gin ON entry_values USING gin(to_tsvector('simple', text_value));

-- MySQL
CREATE FULLTEXT INDEX idx_entry_values_text_ft ON entry_values(text_value);
```

## 七、正则表达式约束

```sql
-- PostgreSQL
CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@...')

-- MySQL 8.0+
CONSTRAINT valid_email CHECK (REGEXP_LIKE(email, '^[A-Za-z0-9._%+-]+@...'))

-- MySQL 5.7（不支持 REGEXP in CHECK，需在应用层验证）
```

## 八、扩展

```sql
-- PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- MySQL / 达梦：不需要，无对应功能
```

## 九、GORM 层面的多数据库适配

项目使用 GORM AutoMigrate，GORM 的 `dialector` 机制可以自动适配不同数据库：

- `gorm.Open(postgres.Open(...), &gorm.Config{})` → PostgreSQL
- `gorm.Open(mysql.Open(...), &gorm.Config{})` → MySQL
- GORM Tags 中的 `type:uuid` / `type:jsonb` 等需要根据 dialect 动态替换

建议：在 GORM 模型层使用 `Dialector` 接口动态生成列类型，而非硬编码 PostgreSQL 类型。
