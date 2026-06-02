// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package graphql

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/contful/contful/openapi/internal/repository"
	"github.com/google/uuid"
)

// Resolver GraphQL 查询解析器
type Resolver struct {
	entryRepo  *repository.EntryRepository
	schemaRepo *repository.ContentSchemaRepository
	fieldRepo  *repository.FieldRepository
}

// NewResolver 创建 Resolver
func NewResolver(
	entryRepo *repository.EntryRepository,
	schemaRepo *repository.ContentSchemaRepository,
	fieldRepo *repository.FieldRepository,
) *Resolver {
	return &Resolver{
		entryRepo:  entryRepo,
		schemaRepo: schemaRepo,
		fieldRepo:  fieldRepo,
	}
}

// contentSchemasResolver 列出所有可用 Content Schema
func (r *Resolver) contentSchemasResolver(ctx context.Context, siteID uuid.UUID) ([]*repository.ContentSchema, error) {
	return r.schemaRepo.ListBySiteID(ctx, siteID)
}

// entryListResolver 分页查询条目列表
func (r *Resolver) entryListResolver(
	c context.Context,
	siteID uuid.UUID,
	schemaSlug string,
	first int,
	after string,
	status string,
	orderBy string,
) ([]map[string]interface{}, error) {
	if first <= 0 || first > 100 {
		first = 20
	}

	// 查找 Content Schema
	cs, err := r.schemaRepo.FindBySlug(c, siteID, schemaSlug)
	if err != nil {
		return nil, fmt.Errorf("content schema '%s' not found", schemaSlug)
	}

	// 获取 Fields
	fields, err := r.fieldRepo.ListByContentSchemaID(c, cs.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load fields: %w", err)
	}

	// 构建查询条件
	var filterStatus string
	switch status {
	case "published":
		filterStatus = "published"
	case "draft":
		filterStatus = "draft"
	default:
		filterStatus = "" // all
	}

	// 查询条目
	entries, err := r.entryRepo.List(c, siteID, cs.ID, &repository.GraphQLFilter{
		Status: filterStatus,
		Order:  orderBy,
		Limit:  first,
		After:  after,
	})
	if err != nil {
		return nil, fmt.Errorf("query entries failed: %w", err)
	}

	// 转换为 GraphQL 结果
	result := make([]map[string]interface{}, 0, len(entries))
	for i := range entries {
		result = append(result, r.entryToMap(&entries[i], fields))
	}
	return result, nil
}

// singleEntryResolver 按 slug 查询单条
func (r *Resolver) singleEntryResolver(c context.Context, siteID uuid.UUID, schemaSlug string, slug string) (map[string]interface{}, error) {
	cs, err := r.schemaRepo.FindBySlug(c, siteID, schemaSlug)
	if err != nil {
		return nil, fmt.Errorf("content schema '%s' not found", schemaSlug)
	}

	fields, err := r.fieldRepo.ListByContentSchemaID(c, cs.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load fields: %w", err)
	}

	entry, err := r.entryRepo.FindBySlug(c, siteID, cs.ID, slug)
	if err != nil {
		return nil, fmt.Errorf("entry not found")
	}

	return r.entryToMap(entry, fields), nil
}

// singleEntryByIDResolver 按 ID 查询单条
func (r *Resolver) singleEntryByIDResolver(c context.Context, siteID uuid.UUID, schemaSlug string, id string) (map[string]interface{}, error) {
	cs, err := r.schemaRepo.FindBySlug(c, siteID, schemaSlug)
	if err != nil {
		return nil, fmt.Errorf("content schema '%s' not found", schemaSlug)
	}

	fields, err := r.fieldRepo.ListByContentSchemaID(c, cs.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load fields: %w", err)
	}

	entryID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid entry id")
	}

	entry, err := r.entryRepo.FindByID(c, entryID)
	if err != nil {
		return nil, fmt.Errorf("entry not found")
	}

	// 验证 entry 属于当前 site 和 schema
	if entry.SiteID != siteID || entry.ContentSchemaID != cs.ID {
		return nil, fmt.Errorf("entry not found")
	}

	return r.entryToMap(entry, fields), nil
}

// entryToMap 将 Entry 转换为 GraphQL 可返回的 map
func (r *Resolver) entryToMap(entry *repository.Entry, fields []repository.Field) map[string]interface{} {
	// 构建 fieldName → EntryValue 映射
	valueMap := make(map[string]*repository.EntryValue)
	for i := range entry.Values {
		valueMap[entry.Values[i].Field.Name] = &entry.Values[i]
	}

	result := map[string]interface{}{
		"_id":            entry.ID.String(),
		"_status":        entry.Status,
		"_createdTime":   entry.CreatedTime,
		"_updatedTime":   entry.UpdatedTime,
		"_publishedTime": entry.PublishedTime,
		"_seoTitle":      entry.SeoTitle,
		"_seoDescription": entry.SeoDescription,
		"_version":       entry.Version,
	}

	// 从 Values 中提取 slug（如果存在 slug 字段）
	for _, v := range entry.Values {
		if v.Field.Name == "slug" {
			if s, ok := v.Value["text"].(string); ok {
				result["_slug"] = s
			}
			break
		}
	}

	// 解析每个自定义字段
	for _, f := range fields {
		v, ok := valueMap[f.Name]
		if !ok {
			result[f.Name] = nil
			continue
		}
		result[f.Name] = extractFieldValue(v, f.FieldType)
	}

	return result
}

// extractFieldValue 从 EntryValue 中提取字段值
func extractFieldValue(v *repository.EntryValue, fieldType string) interface{} {
	if v == nil || v.Value == nil {
		return nil
	}

	switch fieldType {
	case "text", "textarea", "rich_text", "slug":
		if t, ok := v.Value["text"]; ok {
			return t
		}
	case "number":
		if n, ok := v.Value["number"]; ok {
			switch val := n.(type) {
			case float64:
				if val == float64(int64(val)) {
					return int64(val)
				}
				return val
			default:
				return n
			}
		}
	case "boolean":
		if b, ok := v.Value["boolean"]; ok {
			return b
		}
	case "datetime":
		if t, ok := v.Value["datetime"]; ok {
			if s, ok := t.(string); ok {
				if parsed, err := time.Parse(time.RFC3339, s); err == nil {
					return parsed
				}
				return s
			}
		}
	case "media":
		if id, ok := v.Value["media_id"]; ok {
			return map[string]interface{}{"id": id}
		}
		if url, ok := v.Value["url"]; ok {
			return map[string]interface{}{"url": url}
		}
	case "relation":
		if id, ok := v.Value["relation_id"]; ok {
			return map[string]interface{}{"id": id}
		}
	case "color":
		if c, ok := v.Value["color"]; ok {
			return c
		}
	case "json":
		if raw, ok := v.Value["json"]; ok {
			return raw
		}
	default:
		// 返回整个 value map
		return v.Value
	}
	return nil
}

// encodeCursor 创建分页 cursor（base64 编码 entry ID）
func encodeCursor(entryID uuid.UUID) string {
	return base64.StdEncoding.EncodeToString([]byte(entryID.String()))
}

// decodeCursor 解码分页 cursor
func decodeCursor(cursor string) (uuid.UUID, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid cursor")
	}
	return uuid.Parse(string(data))
}

// jsonMarshal helper
func jsonMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
