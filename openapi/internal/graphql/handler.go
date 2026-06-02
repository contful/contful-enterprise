// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package graphql

import (
	"context"
	"fmt"
	"net/http"

	"github.com/contful/contful/openapi/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	gqlHandler "github.com/graphql-go/handler"
)

// GraphQLHandler 创建 Gin GraphQL HTTP Handler
func GraphQLHandler(builder *SchemaBuilder) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Context 获取 siteID
		tc := middleware.GetTokenContext(c)
		if tc == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []gin.H{{"message": "unauthorized"}},
			})
			return
		}

		siteID := tc.SiteID

		// GET 请求显示 GraphiQL Playground（开发环境可用）
		if c.Request.Method == http.MethodGet {
			c.HTML(http.StatusOK, "", nil)
			c.Writer.WriteString(graphiQLHTML)
			return
		}

		// POST 请求执行 GraphQL 查询
		schema, err := builder.Build(siteID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []gin.H{{"message": fmt.Sprintf("failed to build schema: %v", err)}},
			})
			return
		}

		// 读取请求 body
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []gin.H{{"message": "failed to read request body"}},
			})
			return
		}

		// 构建带 siteID 的 context
		ctx := context.WithValue(c.Request.Context(), ctxKeySiteID, siteID)

		// 执行 GraphQL 查询
		result := graphql.Do(graphql.Params{
			Schema:        *schema,
			RequestString: string(body),
			Context:       ctx,
		})

		if len(result.Errors) > 0 {
			c.JSON(http.StatusOK, result)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// PlaygroundHandler 返回 GraphiQL playground（GET /graphql）
func PlaygroundHandler() gin.HandlerFunc {
	h := gqlHandler.New(&gqlHandler.Config{
		Schema:     nil, // 动态 schema，playground 在请求时构建
		Pretty:     true,
		Playground: true,
	})

	return func(c *gin.Context) {
		// Playground HTML
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, graphiQLHTML)
		_ = h
	}
}

// context key 类型
type contextKey string

const ctxKeySiteID contextKey = "siteID"

// GetSiteIDFromContext 从 context 取 siteID
func GetSiteIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxKeySiteID).(uuid.UUID)
	return id, ok
}

// GraphiQL HTML 页面（精简版，支持自定义 endpoint）
const graphiQLHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Contful GraphQL Playground</title>
  <style>
    body { margin: 0; padding: 0; height: 100vh; }
    #graphiql { height: 100vh; }
  </style>
  <link href="https://unpkg.com/graphiql@3.0.0/graphiql.min.css" rel="stylesheet" />
</head>
<body>
  <div id="graphiql">Loading GraphiQL...</div>
  <script crossorigin src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
  <script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
  <script crossorigin src="https://unpkg.com/graphiql@3.0.0/graphiql.min.js"></script>
  <script>
    const url = window.location.origin + window.location.pathname;
    const params = new URLSearchParams(window.location.search);
    const queryParam = params.get('query') || '# Welcome to Contful GraphQL API\n# First query available schemas:\n# { _schemas { slug name } }\n';
    const fetcher = GraphiQL.createFetcher({ url: url });
    const root = React.createElement(GraphiQL, { fetcher: fetcher, defaultQuery: queryParam, defaultEditorToolVisibility: true });
    ReactDOM.render(root, document.getElementById('graphiql'));
  </script>
</body>
</html>`
