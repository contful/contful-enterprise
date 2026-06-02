// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package graphql

import (
	"github.com/graphql-go/graphql"
)

// AssetType 文件资源类型
var AssetType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Asset",
	Fields: graphql.Fields{
		"id":       &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"url":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"filename": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"mimeType": &graphql.Field{Type: graphql.String},
		"size":     &graphql.Field{Type: graphql.Int},
		"width":    &graphql.Field{Type: graphql.Int},
		"height":   &graphql.Field{Type: graphql.Int},
	},
})

// SiteType 站点信息类型
var SiteType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Site",
	Fields: graphql.Fields{
		"id":      &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"name":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"slug":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"locales": &graphql.Field{Type: graphql.NewList(graphql.String)},
	},
})

// PageInfoType Relay 分页信息
var PageInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PageInfo",
	Fields: graphql.Fields{
		"hasNextPage":     &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		"hasPreviousPage": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		"startCursor":     &graphql.Field{Type: graphql.String},
		"endCursor":       &graphql.Field{Type: graphql.String},
		"totalCount":      &graphql.Field{Type: graphql.Int},
	},
})

// ContentSchemaInfoType 内容模型信息（用于 _schemas 查询）
var ContentSchemaInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ContentSchemaInfo",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"name":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"slug":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"description": &graphql.Field{Type: graphql.String},
	},
})
