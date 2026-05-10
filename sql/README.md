# Contful 数据库初始化指南

## 📁 文件结构

```
contful/sql/
├── init_pg.sql   # 数据库初始化脚本（完整版）
└── README.md     # 本文档
```

## 🚀 初始化方式

### 方式一：手动部署（生产环境）

直接执行 `init_pg.sql` 脚本：

```bash
psql -h <host> -U <user> -d <db> -f init_pg.sql
```

**注意事项**：
- 脚本会先删除所有已存在的表、类型、函数（使用 `DROP ... CASCADE`）
- 适用于**全新部署**或**开发环境重置**
- **不适用于生产环境升级**（会丢失数据）

### 方式二：Docker 自动初始化

将 `init_pg.sql` 放到 `docker-entrypoint-initdb.d/` 目录下，PostgreSQL 容器启动时会自动执行。

```dockerfile
COPY init_pg.sql /docker-entrypoint-initdb.d/01-init_pg.sql
```

### 方式三：GORM AutoMigrate（开发环境）

在应用启动时启用 GORM AutoMigrate，自动同步模型变更：

```go
// 在应用初始化时调用
func MigrateDB(db *gorm.DB) error {
    return db.AutoMigrate(
        &model.SystemUser{},
        &model.SystemRole{},
        &model.SystemUserRole{},
        &model.SystemConfig{},
        &model.AuditLog{},
        &model.Site{},
        &model.Schema{},
        &model.Field{},
        &model.Entry{},
        &model.EntryValue{},
        &model.EntryVersion{},
        &model.AssetFolder{},
        &model.Asset{},
        &model.Token{},
    )
}
```

## ⚠️ GORM AutoMigrate 兼容性说明

### ✅ GORM AutoMigrate 可以处理：

- 创建表（如果不存在）
- 添加新列
- 创建基础索引（`gorm:"index"` 标签）
- 创建唯一约束（`gorm:"uniqueIndex"` 标签）

### ❌ GORM AutoMigrate 无法处理：

- 创建 ENUM 类型（需要预先执行 `CREATE TYPE`）
- 创建触发器/函数
- 添加表注释和列注释（COMMENT）
- 创建复杂索引（部分索引、GIN 索引等）
- 修改列类型
- 删除未使用的列

### 🔧 解决方案

1. **首次部署**：执行 `init_pg.sql` 创建完整 schema
2. **后续升级**：
   - 简单变更（新增列、新增索引）：依赖 GORM AutoMigrate 自动处理
   - 复杂变更（修改列类型、添加 ENUM 值、创建触发器）：需要手动编写 SQL 脚本

## 🧪 验证数据库结构

运行以下查询，检查 GORM 模型与数据库表结构是否一致：

```sql
-- 检查表是否存在
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;

-- 检查列类型是否匹配
SELECT 
    table_name, 
    column_name, 
    data_type, 
    udt_name 
FROM information_schema.columns 
WHERE table_schema = 'public' 
ORDER BY table_name, column_name;
```

## 🚨 常见问题

### Q1: GORM AutoMigrate 报错 "type XXX does not exist"

**原因**：GORM 不会自动创建 ENUM 类型。

**解决**：确保先执行 `init_pg.sql`（包含 `CREATE TYPE` 语句）。

### Q2: 如何安全升级生产环境？

**推荐流程**：

1. 在测试环境验证 SQL 变更脚本
2. 备份生产数据库
3. 编写增量升级脚本（不要使用 `init_pg.sql` 的完整重置方式）
4. 在低峰期执行升级
5. 准备回滚方案

**示例增量升级脚本**：

```sql
-- 升级脚本示例：添加新列（幂等）
ALTER TABLE system_users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
ALTER TABLE system_users ADD COLUMN IF NOT EXISTS department VARCHAR(100);

-- 添加注释
COMMENT ON COLUMN system_users.phone IS '用户手机号';
COMMENT ON COLUMN system_users.department IS '用户所属部门';
```

### Q3: 如何回滚变更？

**方案**：

1. 手动编写回滚脚本（推荐）
2. 从备份恢复

**示例回滚脚本**：

```sql
-- 回滚上面的升级：删除新增的列
ALTER TABLE system_users DROP COLUMN IF EXISTS phone;
ALTER TABLE system_users DROP COLUMN IF EXISTS department;
```

## 📝 init_pg.sql 文件说明

### 文件特点

1. **幂等设计**：使用 `DROP ... IF EXISTS` 和 `CASCADE` 关键字，可重复执行
2. **完整 schema**：包含所有的表、索引、ENUM 类型
3. **GORM 兼容**：表名、字段名、类型与 GORM 模型完全一致
4. **破坏性操作**：会删除所有现有数据，**仅用于全新部署**

### 文件结构

```sql
-- 1. 启用扩展（uuid-ossp, pgcrypto）
-- 2. 清理已有对象（表、类型、函数、触发器）
-- 3. 创建 ENUM 类型
-- 4. 创建所有表结构
-- 5. 创建基础索引
```

### 高级特性（手动添加）

`init_pg.sql` 不包含以下高级特性，需要时可手动添加：

1. **触发器函数**：自动更新 `updated_at` 字段
2. **触发器**：为所有表添加自动更新 `updated_at`
3. **复杂索引**：组合索引、GIN 全文搜索索引、部分索引
4. **数据完整性签名**：`data_signature` 列
5. **审计日志防篡改触发器**
6. **表注释和列注释**

## 📚 参考资料

- [GORM AutoMigrate 官方文档](https://gorm.io/docs/migration.html)
- [PostgreSQL CREATE TYPE 文档](https://www.postgresql.org/docs/current/sql-createtype.html)
- [PostgreSQL 触发器文档](https://www.postgresql.org/docs/current/triggers.html)
- [PostgreSQL DROP 语句文档](https://www.postgresql.org/docs/current/sql-drop-table.html)
