// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package graphql

import (
	"fmt"
	"sync"

	"github.com/contful/contful/openapi/internal/repository"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
)

// SchemaBuilder 动态 GraphQL Schema 构建器
type SchemaBuilder struct {
	schemaRepo *repository.ContentSchemaRepository
	fieldRepo  *repository.FieldRepository
	resolver   *Resolver

	mu    sync.RWMutex
	cache map[uuid.UUID]graphql.Schema // siteID → Schema 缓存
}

// NewSchemaBuilder 创建 SchemaBuilder
func NewSchemaBuilder(
	schemaRepo *repository.ContentSchemaRepository,
	fieldRepo *repository.FieldRepository,
	resolver *Resolver,
) *SchemaBuilder {
	return &SchemaBuilder{
		schemaRepo: schemaRepo,
		fieldRepo:  fieldRepo,
		resolver:   resolver,
		cache:      make(map[uuid.UUID]graphql.Schema),
	}
}

// BuildQuery 构建 Query 类型
func (b *SchemaBuilder) BuildQuery(siteID uuid.UUID) (*graphql.Object, error) {
	// 读取站点下所有 Content Schema
	schemas, err := b.schemaRepo.ListBySiteID(nil, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to load content schemas: %w", err)
	}

	queryFields := graphql.Fields{
		"_schemas": &graphql.Field{
			Type: graphql.NewList(ContentSchemaInfoType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				result := make([]map[string]interface{}, 0, len(schemas))
				for _, s := range schemas {
					result = append(result, map[string]interface{}{
						"id":          s.ID.String(),
						"name":        s.Name,
						"slug":        s.Slug,
						"description": s.Description,
					})
				}
				return result, nil
			},
		},
	}

	// 为每个 Content Schema 创建查询字段
	for _, cs := range schemas {
		if err := b.addContentSchemaQuery(queryFields, siteID, cs); err != nil {
			// 跳过无法构建的 Schema，不中断整体
			continue
		}
	}

	return graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: queryFields,
	}), nil
}

// addContentSchemaQuery 为单个 Content Schema 添加查询字段
func (b *SchemaBuilder) addContentSchemaQuery(
	queryFields graphql.Fields,
	siteID uuid.UUID,
	cs *repository.ContentSchema,
) error {
	// 读取 Fields
	fields, err := b.fieldRepo.ListByContentSchemaID(nil, cs.ID)
	if err != nil {
		return fmt.Errorf("failed to load fields for %s: %w", cs.Slug, err)
	}

	// 构建 Object Type
	contentType := b.buildContentType(cs, fields)
	connectionType := b.buildConnectionType(contentType.Name(), contentType)

	slug := cs.Slug

	// 列表查询: <slug>(first, after, status, orderBy): <Type>Connection!
	queryFields[slug] = &graphql.Field{
		Type: graphql.NewNonNull(connectionType),
		Args: graphql.FieldConfigArgument{
			"first":   &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
			"after":   &graphql.ArgumentConfig{Type: graphql.String},
			"status":  &graphql.ArgumentConfig{Type: graphql.String, DefaultValue: "published"},
			"orderBy": &graphql.ArgumentConfig{Type: graphql.String, DefaultValue: "createdAt_DESC"},
			"limit":   &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			first := getIntArg(p.Args, "first", 20)
			after := getStringArg(p.Args, "after", "")
			status := getStringArg(p.Args, "status", "published")

			var orderBy string
			if ob, ok := p.Args["orderBy"]; ok && ob != nil {
				orderBy = fmt.Sprintf("%v", ob)
			} else {
				orderBy = "created_time DESC"
			}

			entries, err := b.resolver.entryListResolver(p.Context, siteID, slug, first, after, status, orderBy)
			if err != nil {
				return nil, err
			}

			hasNext := len(entries) > first
			if hasNext {
				entries = entries[:first]
			}

			var endCursor string
			if len(entries) > 0 {
				if idStr, ok := entries[len(entries)-1]["_id"].(string); ok {
					if id, parseErr := uuid.Parse(idStr); parseErr == nil {
						endCursor = encodeCursor(id)
					}
				}
			}

			return map[string]interface{}{
				"edges":     entries,
				"pageInfo": map[string]interface{}{
					"hasNextPage":     hasNext,
					"hasPreviousPage": after != "",
					"startCursor":     "",
					"endCursor":       endCursor,
				},
			}, nil
		},
	}

	// 单条按 slug: <slug>BySlug(slug: String!): <Type>
	queryFields[slug+"BySlug"] = &graphql.Field{
		Type: contentType,
		Args: graphql.FieldConfigArgument{
			"slug": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			entrySlug := getStringArg(p.Args, "slug", "")
			if entrySlug == "" {
				return nil, fmt.Errorf("slug is required")
			}
			return b.resolver.singleEntryResolver(p.Context, siteID, slug, entrySlug)
		},
	}

	// 单条按 ID: <slug>ById(id: ID!): <Type>
	queryFields[slug+"ById"] = &graphql.Field{
		Type: contentType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			entryID := getStringArg(p.Args, "id", "")
			if entryID == "" {
				return nil, fmt.Errorf("id is required")
			}
			return b.resolver.singleEntryByIDResolver(p.Context, siteID, slug, entryID)
		},
	}

	return nil
}

// buildContentType 为 Content Schema 构建 GraphQL Object Type
func (b *SchemaBuilder) buildContentType(cs *repository.ContentSchema, fields []repository.Field) *graphql.Object {
	gqlFields := graphql.Fields{
		"_id":             &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"_slug":           &graphql.Field{Type: graphql.String},
		"_status":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"_createdTime":    &graphql.Field{Type: graphql.DateTime},
		"_updatedTime":    &graphql.Field{Type: graphql.DateTime},
		"_publishedTime":  &graphql.Field{Type: graphql.DateTime},
		"_seoTitle":       &graphql.Field{Type: graphql.String},
		"_seoDescription": &graphql.Field{Type: graphql.String},
		"_version":        &graphql.Field{Type: graphql.Int},
	}

	// 添加自定义字段
	for _, f := range fields {
		gqlFields[f.Name] = &graphql.Field{
			Type: mapFieldType(f.FieldType),
		}
	}

	return graphql.NewObject(graphql.ObjectConfig{
		Name:   toPascalCase(cs.Slug),
		Fields: gqlFields,
	})
}

// buildConnectionType 构建 Relay Connection 类型
func (b *SchemaBuilder) buildConnectionType(name string, edgeType *graphql.Object) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: name + "Connection",
		Fields: graphql.Fields{
			"edges": &graphql.Field{
				Type: graphql.NewList(edgeType),
			},
			"pageInfo": &graphql.Field{
				Type: graphql.NewNonNull(PageInfoType),
			},
			"totalCount": &graphql.Field{
				Type: graphql.Int,
			},
		},
	})
}

// mapFieldType 将 Contful FieldType 映射到 GraphQL 类型
func mapFieldType(fieldType string) graphql.Output {
	switch fieldType {
	case "number":
		return graphql.Float
	case "boolean":
		return graphql.Boolean
	case "datetime":
		return graphql.DateTime
	case "media":
		return AssetType
	case "json":
		return graphql.String // JSON 序列化为 string
	case "color":
		return graphql.String
	default: // text, textarea, rich_text, slug, relation, repeater 等
		return graphql.String
	}
}

// Build 构建完整 Schema
func (b *SchemaBuilder) Build(siteID uuid.UUID) (*graphql.Schema, error) {
	// 检查缓存
	b.mu.RLock()
	cached, ok := b.cache[siteID]
	b.mu.RUnlock()
	if ok {
		return &cached, nil
	}

	queryType, err := b.BuildQuery(siteID)
	if err != nil {
		return nil, err
	}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// 写入缓存
	b.mu.Lock()
	b.cache[siteID] = schema
	b.mu.Unlock()

	return &schema, nil
}

// InvalidateCache 清除指定站点缓存（Content Schema 变更后调用）
func (b *SchemaBuilder) InvalidateCache(siteID uuid.UUID) {
	b.mu.Lock()
	delete(b.cache, siteID)
	b.mu.Unlock()
}

// toPascalCase 将 snake_case/kebab-case 转为 PascalCase
func toPascalCase(s string) string {
	result := make([]byte, 0, len(s))
	capitalize := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' || c == '-' {
			capitalize = true
			continue
		}
		if capitalize {
			if c >= 'a' && c <= 'z' {
				c -= 32
			}
			capitalize = false
		}
		result = append(result, c)
	}
	return string(result)
}

// getStringArg 安全获取 string 参数
func getStringArg(args map[string]interface{}, key, defaultVal string) string {
	v, ok := args[key]
	if !ok || v == nil {
		return defaultVal
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

// getIntArg 安全获取 int 参数
func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	v, ok := args[key]
	if !ok || v == nil {
		return defaultVal
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	default:
		return defaultVal
	}
}
