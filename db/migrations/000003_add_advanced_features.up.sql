-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Migration 000003: Add advanced database features
-- =============================================================================

-- =============================================================================
-- 1. 触发器函数：自动更新 updated_at
-- =============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS '自动更新 updated_at 时间戳的触发器函数';

-- =============================================================================
-- 2. 触发器：为所有表添加自动更新 updated_at
-- =============================================================================

CREATE TRIGGER update_system_users_updated_at BEFORE UPDATE ON system_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_roles_updated_at BEFORE UPDATE ON system_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_user_roles_updated_at BEFORE UPDATE ON system_user_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_config_updated_at BEFORE UPDATE ON system_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_audit_logs_updated_at BEFORE UPDATE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sites_updated_at BEFORE UPDATE ON sites
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_schemas_updated_at BEFORE UPDATE ON schemas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_fields_updated_at BEFORE UPDATE ON fields
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entries_updated_at BEFORE UPDATE ON entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entry_values_updated_at BEFORE UPDATE ON entry_values
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entry_versions_updated_at BEFORE UPDATE ON entry_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_asset_folders_updated_at BEFORE UPDATE ON asset_folders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_assets_updated_at BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tokens_updated_at BEFORE UPDATE ON tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 3. 复杂索引（GORM AutoMigrate 无法创建）
-- =============================================================================

-- 组合索引
CREATE INDEX IF NOT EXISTS idx_entries_list ON entries(schema_id, locale, status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_entries_site_locale ON entries(site_id, locale, status);
CREATE INDEX IF NOT EXISTS idx_assets_site_type ON assets(site_id, type);
CREATE INDEX IF NOT EXISTS idx_assets_site_created ON assets(site_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_assets_path ON assets(path);
CREATE INDEX IF NOT EXISTS idx_audit_logs_site_user_time ON audit_logs(site_id, user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_category_time ON audit_logs(category, created_at DESC);

-- GIN 索引（全文搜索）
CREATE INDEX IF NOT EXISTS idx_entry_values_text_gin ON entry_values USING gin(to_tsvector('simple', text_value)) WHERE text_value IS NOT NULL;

-- =============================================================================
-- 4. 数据完整性签名（M3）
-- =============================================================================

-- entries 表新增签名列
ALTER TABLE entries ADD COLUMN IF NOT EXISTS data_signature JSONB;
COMMENT ON COLUMN entries.data_signature IS '数据完整性签名: {"alg":"HMAC-SHA256","created_at":"...","payload_hash":"...","signature":"..."}';

-- entry_values 表新增签名列（联动签名字段值）
ALTER TABLE entry_values ADD COLUMN IF NOT EXISTS data_signature JSONB;
COMMENT ON COLUMN entry_values.data_signature IS '字段值完整性签名';

-- assets 表新增签名列
ALTER TABLE assets ADD COLUMN IF NOT EXISTS data_signature JSONB;
COMMENT ON COLUMN assets.data_signature IS '媒体资源元信息完整性签名';

-- audit_logs 表新增签名列（NOT NULL，历史数据默认空对象）
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS data_signature JSONB NOT NULL DEFAULT '{}';
COMMENT ON COLUMN audit_logs.data_signature IS '审计日志完整性签名，防篡改';

-- schemas 表新增签名开关
ALTER TABLE schemas ADD COLUMN IF NOT EXISTS signature_enabled BOOLEAN NOT NULL DEFAULT FALSE;
COMMENT ON COLUMN schemas.signature_enabled IS '是否启用数据签名（该内容类型下的条目自动签名）';

-- =============================================================================
-- 5. 审计日志防篡改触发器
-- =============================================================================

CREATE OR REPLACE FUNCTION prevent_audit_log_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION '审计日志禁止更新，仅允许 INSERT 和软删除';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS prevent_audit_logs_update ON audit_logs;
CREATE TRIGGER prevent_audit_logs_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_update();

-- =============================================================================
-- 6. RBAC 预置系统角色数据（幂等可重复执行）
-- =============================================================================

-- 1. 超级管理员角色（拥有全部系统级权限）
INSERT INTO system_roles (id, name, description, is_system, permissions, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000101'::uuid,
    'Super Admin',
    '系统超级管理员，拥有全部权限',
    TRUE,
    '["system:users:read","system:users:write","system:users:delete",
      "system:sites:read","system:sites:write","system:sites:delete",
      "system:tokens:read","system:tokens:write","system:tokens:delete",
      "system:settings:read","system:settings:write",
      "system:audit:read","system:audit:export",
      "system:roles:read","system:roles:write","system:roles:delete"]'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM system_roles WHERE id = '00000000-0000-0000-0000-000000000101'::uuid);

-- 2. 审计角色（仅可查看和导出审计日志）
INSERT INTO system_roles (id, name, description, is_system, permissions, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000103'::uuid,
    'Auditor',
    '审计人员，仅可查看和导出审计日志',
    TRUE,
    '["system:users:read","system:sites:read","system:tokens:read",
      "system:settings:read","system:audit:read","system:audit:export"]'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM system_roles WHERE id = '00000000-0000-0000-0000-000000000103'::uuid);

-- =============================================================================
-- 7. 表注释和列注释
-- =============================================================================

-- system_users 表注释
COMMENT ON TABLE system_users IS '系统用户表：存放所有用户账户，用户全局管理，权限通过 system_user_roles 关联';
COMMENT ON COLUMN system_users.id IS '用户唯一标识符(UUID)';
COMMENT ON COLUMN system_users.email IS '用户邮箱（全局唯一）';
COMMENT ON COLUMN system_users.password_hash IS 'bcrypt 加密的密码哈希';
COMMENT ON COLUMN system_users.nickname IS '用户昵称';
COMMENT ON COLUMN system_users.avatar_url IS '头像 URL';
COMMENT ON COLUMN system_users.status IS '账户状态：active/inactive/suspended';
COMMENT ON COLUMN system_users.is_super_admin IS '是否超级管理员（拥有全部权限，不受站点限制）';
COMMENT ON COLUMN system_users.last_login_at IS '最后登录时间';
COMMENT ON COLUMN system_users.last_login_ip IS '最后登录 IP 地址';
COMMENT ON COLUMN system_users.created_at IS '创建时间';
COMMENT ON COLUMN system_users.updated_at IS '更新时间';
COMMENT ON COLUMN system_users.deleted_at IS '软删除时间（非空表示已删除）';

-- system_roles 表注释
COMMENT ON TABLE system_roles IS '系统角色表：系统级角色定义，系统级角色不可删除';
COMMENT ON COLUMN system_roles.id IS '角色唯一标识符';
COMMENT ON COLUMN system_roles.name IS '角色名称（全局唯一）';
COMMENT ON COLUMN system_roles.description IS '角色描述';
COMMENT ON COLUMN system_roles.is_system IS '是否系统角色（系统角色不可删除）';
COMMENT ON COLUMN system_roles.permissions IS '权限列表 JSON：["user:read", "user:write", ...]';
COMMENT ON COLUMN system_roles.created_at IS '创建时间';
COMMENT ON COLUMN system_roles.updated_at IS '更新时间';
COMMENT ON COLUMN system_roles.deleted_at IS '软删除时间';

-- system_user_roles 表注释
COMMENT ON TABLE system_user_roles IS '系统用户-角色关联表：多对多关系，一个用户可拥有多个系统角色';
COMMENT ON COLUMN system_user_roles.id IS '关联记录唯一标识符';
COMMENT ON COLUMN system_user_roles.user_id IS '用户 ID';
COMMENT ON COLUMN system_user_roles.role_id IS '系统角色 ID';
COMMENT ON COLUMN system_user_roles.created_at IS '创建时间';

-- system_config 表注释
COMMENT ON TABLE system_config IS '系统配置表：存储全局配置项（密码策略、系统名称、Logo 等）';
COMMENT ON COLUMN system_config.id IS '配置项唯一标识符';
COMMENT ON COLUMN system_config.config_key IS '配置键名（唯一，如 password_expire_days、site_name、logo_url）';
COMMENT ON COLUMN system_config.config_value IS '配置值（根据 value_type 解析）';
COMMENT ON COLUMN system_config.value_type IS '值类型：string/number/boolean/json';
COMMENT ON COLUMN system_config.description IS '配置项描述';
COMMENT ON COLUMN system_config.is_public IS '是否公开（公开配置可未授权访问）';
COMMENT ON COLUMN system_config.created_at IS '创建时间';
COMMENT ON COLUMN system_config.updated_at IS '更新时间';

-- audit_logs 表注释
COMMENT ON TABLE audit_logs IS '审计日志表：记录所有操作行为';
COMMENT ON COLUMN audit_logs.id IS '日志唯一标识符';
COMMENT ON COLUMN audit_logs.site_id IS '关联站点 ID（全局操作时为空）';
COMMENT ON COLUMN audit_logs.user_id IS '操作用户 ID';
COMMENT ON COLUMN audit_logs.action IS '操作动作名称';
COMMENT ON COLUMN audit_logs.resource_type IS '资源类型';
COMMENT ON COLUMN audit_logs.resource_id IS '资源 ID';
COMMENT ON COLUMN audit_logs.level IS '日志级别：debug/info/warn/error';
COMMENT ON COLUMN audit_logs.category IS '操作类别：auth/content/media/settings/user/system';
COMMENT ON COLUMN audit_logs.details IS '操作详情 JSON';
COMMENT ON COLUMN audit_logs.ip_address IS '客户端 IP';
COMMENT ON COLUMN audit_logs.user_agent IS '客户端 User-Agent';
COMMENT ON COLUMN audit_logs.created_at IS '日志创建时间';

-- sites 表注释
COMMENT ON TABLE sites IS '站点表：每个站点是一个独立的 CMS 实例';
COMMENT ON COLUMN sites.id IS '站点唯一标识符';
COMMENT ON COLUMN sites.name IS '站点名称';
COMMENT ON COLUMN sites.slug IS '站点别名（URL 中使用）';
COMMENT ON COLUMN sites.description IS '站点描述';
COMMENT ON COLUMN sites.site_url IS '站点访问地址（前端展示端点）';
COMMENT ON COLUMN sites.locale IS '站点默认语言（如 zh-CN、en-US）';
COMMENT ON COLUMN sites.timezone IS '站点时区（如 Asia/Shanghai、America/New_York）';
COMMENT ON COLUMN sites.seo_title IS '站点 SEO 标题';
COMMENT ON COLUMN sites.seo_description IS '站点 SEO 描述';
COMMENT ON COLUMN sites.seo_keywords IS '站点 SEO 关键词';
COMMENT ON COLUMN sites.settings IS '动态配置 JSON（主题、第三方集成等）';
COMMENT ON COLUMN sites.is_active IS '是否激活';
COMMENT ON COLUMN sites.created_by IS '创建者用户 ID';
COMMENT ON COLUMN sites.created_at IS '创建时间';
COMMENT ON COLUMN sites.updated_at IS '更新时间';
COMMENT ON COLUMN sites.deleted_at IS '软删除时间';

-- schemas 表注释
COMMENT ON TABLE schemas IS '内容模型表：定义内容的结构（类似 WordPress 的自定义内容类型）';
COMMENT ON COLUMN schemas.id IS '内容模型唯一标识符';
COMMENT ON COLUMN schemas.site_id IS '所属站点';
COMMENT ON COLUMN schemas.name IS '模型名称';
COMMENT ON COLUMN schemas.slug IS '模型别名（API 中使用）';
COMMENT ON COLUMN schemas.description IS '模型描述';
COMMENT ON COLUMN schemas.kind IS '类型：collection=集合、single=单页';
COMMENT ON COLUMN schemas.display_config IS '前端显示配置 JSON';
COMMENT ON COLUMN schemas.api_config IS 'API 配置 JSON：publicRead/publicWrite 等';
COMMENT ON COLUMN schemas.preview_config IS '预览配置 JSON';
COMMENT ON COLUMN schemas.versioning_enabled IS '是否启用版本历史';
COMMENT ON COLUMN schemas.draft_autosave_interval IS '草稿自动保存间隔（秒）';
COMMENT ON COLUMN schemas.is_active IS '是否启用';
COMMENT ON COLUMN schemas.sort_order IS '排序权重';
COMMENT ON COLUMN schemas.created_by IS '创建者';
COMMENT ON COLUMN schemas.created_at IS '创建时间';
COMMENT ON COLUMN schemas.updated_at IS '更新时间';
COMMENT ON COLUMN schemas.deleted_at IS '软删除时间';

-- fields 表注释
COMMENT ON TABLE fields IS '字段表：内容模型的字段定义';
COMMENT ON COLUMN fields.id IS '字段唯一标识符';
COMMENT ON COLUMN fields.schema_id IS '所属内容模型';
COMMENT ON COLUMN fields.name IS '字段名（代码中使用）';
COMMENT ON COLUMN fields.label IS '字段显示标签';
COMMENT ON COLUMN fields.description IS '字段描述';
COMMENT ON COLUMN fields.field_type IS '字段类型：text/rich_text/number/boolean/date/datetime/email/url/json/media/relation/enum/password';
COMMENT ON COLUMN fields.config IS '字段配置 JSON：maxLength/minValue/options 等';
COMMENT ON COLUMN fields.validation IS '验证规则 JSON：required/pattern/min/max 等';
COMMENT ON COLUMN fields.display IS '显示配置 JSON：placeholder/readOnly 等';
COMMENT ON COLUMN fields.default_value IS '默认值 JSON';
COMMENT ON COLUMN fields.sort_order IS '排序权重';
COMMENT ON COLUMN fields.conditional_display IS '条件显示 JSON';
COMMENT ON COLUMN fields.created_at IS '创建时间';
COMMENT ON COLUMN fields.updated_at IS '更新时间';
COMMENT ON COLUMN fields.deleted_at IS '软删除时间';

-- entries 表注释
COMMENT ON TABLE entries IS '内容条目表：实际的内容数据';
COMMENT ON COLUMN entries.id IS '条目唯一标识符';
COMMENT ON COLUMN entries.schema_id IS '所属内容模型';
COMMENT ON COLUMN entries.site_id IS '所属站点';
COMMENT ON COLUMN entries.locale IS '语言/地区代码（如 zh-CN、en-US）';
COMMENT ON COLUMN entries.status IS '状态：draft/published/archived';
COMMENT ON COLUMN entries.version IS '当前版本号';
COMMENT ON COLUMN entries.version_history IS '版本历史 JSON';
COMMENT ON COLUMN entries.published_at IS '发布时间';
COMMENT ON COLUMN entries.published_by IS '发布者';
COMMENT ON COLUMN entries.relations IS '关联条目 JSON';
COMMENT ON COLUMN entries.seo_title IS 'SEO 标题';
COMMENT ON COLUMN entries.seo_description IS 'SEO 描述';
COMMENT ON COLUMN entries.seo_keywords IS 'SEO 关键词';
COMMENT ON COLUMN entries.sort_weight IS '排序权重';
COMMENT ON COLUMN entries.created_by IS '创建者';
COMMENT ON COLUMN entries.created_at IS '创建时间';
COMMENT ON COLUMN entries.updated_at IS '更新时间';
COMMENT ON COLUMN entries.deleted_at IS '软删除时间';

-- entry_values 表注释
COMMENT ON TABLE entry_values IS '条目值表：字段的实际值（一行对应一个字段的值）';
COMMENT ON COLUMN entry_values.id IS '值唯一标识符';
COMMENT ON COLUMN entry_values.entry_id IS '所属条目';
COMMENT ON COLUMN entry_values.field_id IS '所属字段';
COMMENT ON COLUMN entry_values.value IS '值 JSONB（通用存储）';
COMMENT ON COLUMN entry_values.text_value IS '文本值（索引用）';
COMMENT ON COLUMN entry_values.number_value IS '数值（索引用）';
COMMENT ON COLUMN entry_values.bool_value IS '布尔值（索引用）';
COMMENT ON COLUMN entry_values.date_value IS '日期值（索引用）';
COMMENT ON COLUMN entry_values.datetime_value IS '时间戳值（索引用）';
COMMENT ON COLUMN entry_values.created_at IS '创建时间';
COMMENT ON COLUMN entry_values.updated_at IS '更新时间';

-- entry_versions 表注释
COMMENT ON TABLE entry_versions IS '条目版本表：内容的历史版本记录';
COMMENT ON COLUMN entry_versions.id IS '版本唯一标识符';
COMMENT ON COLUMN entry_versions.entry_id IS '所属条目';
COMMENT ON COLUMN entry_versions.version IS '版本号';
COMMENT ON COLUMN entry_versions.values_snapshot IS '值快照 JSON';
COMMENT ON COLUMN entry_versions.created_by IS '创建者';
COMMENT ON COLUMN entry_versions.created_at IS '创建时间';
COMMENT ON COLUMN entry_versions.change_summary IS '变更摘要';

-- asset_folders 表注释
COMMENT ON TABLE asset_folders IS '资产文件夹表：媒体库文件夹结构';
COMMENT ON COLUMN asset_folders.id IS '文件夹唯一标识符';
COMMENT ON COLUMN asset_folders.site_id IS '所属站点';
COMMENT ON COLUMN asset_folders.parent_id IS '父文件夹';
COMMENT ON COLUMN asset_folders.name IS '文件夹名称';
COMMENT ON COLUMN asset_folders.slug IS '文件夹别名';
COMMENT ON COLUMN asset_folders.path IS '完整路径';
COMMENT ON COLUMN asset_folders.sort_order IS '排序权重';
COMMENT ON COLUMN asset_folders.created_by IS '创建者';
COMMENT ON COLUMN asset_folders.created_at IS '创建时间';
COMMENT ON COLUMN asset_folders.updated_at IS '更新时间';
COMMENT ON COLUMN asset_folders.deleted_at IS '软删除时间';

-- assets 表注释
COMMENT ON TABLE assets IS '资产表：图片/视频/文档等媒体文件';
COMMENT ON COLUMN assets.id IS '资产唯一标识符';
COMMENT ON COLUMN assets.site_id IS '所属站点';
COMMENT ON COLUMN assets.folder_id IS '所属文件夹';
COMMENT ON COLUMN assets.uuid IS '业务 UUID（用于引用）';
COMMENT ON COLUMN assets.name IS '显示名称';
COMMENT ON COLUMN assets.original_name IS '原始文件名';
COMMENT ON COLUMN assets.slug IS 'URL 别名';
COMMENT ON COLUMN assets.type IS '资源类型';
COMMENT ON COLUMN assets.mime_type IS 'MIME 类型';
COMMENT ON COLUMN assets.extension IS '文件扩展名';
COMMENT ON COLUMN assets.size IS '文件大小（字节）';
COMMENT ON COLUMN assets.width IS '宽度（图片/视频）';
COMMENT ON COLUMN assets.height IS '高度（图片/视频）';
COMMENT ON COLUMN assets.duration IS '时长（音视频，秒）';
COMMENT ON COLUMN assets.path IS '存储路径';
COMMENT ON COLUMN assets.url IS '访问 URL';
COMMENT ON COLUMN assets.thumbnail_url IS '缩略图 URL';
COMMENT ON COLUMN assets.alt IS 'Alt 文本';
COMMENT ON COLUMN assets.title IS '标题';
COMMENT ON COLUMN assets.caption IS '图注';
COMMENT ON COLUMN assets.alt_text IS '备用描述';
COMMENT ON COLUMN assets.description IS '描述';
COMMENT ON COLUMN assets.tags IS '标签列表';
COMMENT ON COLUMN assets.metadata IS '元数据 JSON：EXIF 等';
COMMENT ON COLUMN assets.visibility IS '可见性：public/private';
COMMENT ON COLUMN assets.file_hash IS '文件哈希（SHA256）';
COMMENT ON COLUMN assets.disk IS '存储驱动：local/s3/oss/cos 等';
COMMENT ON COLUMN assets.download_count IS '下载次数';
COMMENT ON COLUMN assets.used_count IS '引用次数';
COMMENT ON COLUMN assets.created_by IS '上传者';
COMMENT ON COLUMN assets.created_at IS '上传时间';
COMMENT ON COLUMN assets.updated_at IS '更新时间';
COMMENT ON COLUMN assets.deleted_at IS '软删除时间';

-- tokens 表注释
COMMENT ON TABLE tokens IS 'API Token 表：用于 Open API 认证';
COMMENT ON COLUMN tokens.id IS 'Token 唯一标识符';
COMMENT ON COLUMN tokens.site_id IS '所属站点';
COMMENT ON COLUMN tokens.name IS 'Token 名称';
COMMENT ON COLUMN tokens.description IS 'Token 描述';
COMMENT ON COLUMN tokens.token_prefix IS 'Token 前缀（显示用）';
COMMENT ON COLUMN tokens.token_hash IS 'Token 哈希（验证用）';
COMMENT ON COLUMN tokens.permissions IS '权限范围 JSON';
COMMENT ON COLUMN tokens.rate_limits IS '速率限制 JSON';
COMMENT ON COLUMN tokens.usage IS '使用统计 JSON';
COMMENT ON COLUMN tokens.expires_at IS '过期时间';
COMMENT ON COLUMN tokens.status IS '状态：active/expired/revoked';
COMMENT ON COLUMN tokens.last_used_at IS '最后使用时间';
COMMENT ON COLUMN tokens.created_by IS '创建者';
COMMENT ON COLUMN tokens.created_at IS '创建时间';
COMMENT ON COLUMN tokens.updated_at IS '更新时间';
COMMENT ON COLUMN tokens.deleted_at IS '软删除时间';

-- ENUM 类型注释
COMMENT ON TYPE user_status IS '用户账户状态：active=正常、inactive=未激活、suspended=已停用';
COMMENT ON TYPE entry_status IS '内容条目状态：draft=草稿、published=已发布、archived=已归档';
COMMENT ON TYPE content_type_kind IS '内容模型类型：collection=集合（如文章列表）、single=单页（如关于页面）';
COMMENT ON TYPE field_type IS '字段类型：text=文本、rich_text=富文本、number=数字、boolean=布尔、date=日期、datetime=时间、email=邮箱、url=链接、json=JSON、media=媒体、relation=关联、enum=枚举、password=密码';
COMMENT ON TYPE asset_type IS '资源类型：image=图片、video=视频、audio=音频、document=文档、file=文件';
COMMENT ON TYPE asset_visibility IS '资源可见性：public=公开、private=私有';
COMMENT ON TYPE asset_status IS '资源状态：active=正常、inactive=禁用、deleted=已删除';
COMMENT ON TYPE token_status IS 'API Token 状态：active=有效、expired=已过期、revoked=已撤销';
COMMENT ON TYPE audit_level IS '审计日志级别：debug=调试、info=信息、warn=警告、error=错误';
COMMENT ON TYPE audit_type IS '审计日志类别：auth=认证、content=内容、media=媒体、settings=设置、user=用户、system=系统';
