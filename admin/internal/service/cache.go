// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Redis 缓存键命名规范（Admin 与 OpenAPI 共用）
// ─────────────────────────────────────────────────────
//
//	contful:content:list:{siteID}:{slug}:{locale}:{sortField}:{sortOrder}:{page}:{pageSize}
//	contful:content:detail:{siteID}:{slug}:{entryID}
//	contful:config:{siteID}:{key}
//
// 清除策略:
//
//	全局 → SCAN contful:*:*             → DEL
//	站点 → SCAN contful:*:{siteID}:*    → DEL
//	模型 → SCAN contful:content:*:{siteID}:{slug}:* → DEL

const (
	// KeyAllPattern   全局缓存匹配模式
	KeyAllPattern = "contful:*:*"
	// KeySitePattern 站点缓存匹配模式
	KeySitePattern = "contful:*:%s:*"
	// KeySchemaPattern 模型缓存匹配模式
	KeySchemaPattern = "contful:content:*:%s:%s:*"
)

// CacheService Admin 缓存服务（用于清除 OpenAPI 缓存）
type CacheService struct {
	rdb *redis.Client
}

// NewCacheService 创建 CacheService
func NewCacheService(rdb *redis.Client) *CacheService {
	return &CacheService{rdb: rdb}
}

// InvalidateAll 清除所有 contful 前缀的缓存
func (s *CacheService) InvalidateAll(ctx context.Context) (int64, error) {
	return s.scanAndDel(ctx, KeyAllPattern)
}

// InvalidateSite 清除指定站点的所有缓存
func (s *CacheService) InvalidateSite(ctx context.Context, siteID string) (int64, error) {
	return s.scanAndDel(ctx, fmt.Sprintf(KeySitePattern, siteID))
}

// InvalidateSchema 清除指定站点+模型的缓存
func (s *CacheService) InvalidateSchema(ctx context.Context, siteID, slug string) (int64, error) {
	return s.scanAndDel(ctx, fmt.Sprintf(KeySchemaPattern, siteID, slug))
}

// scanAndDel SCAN 匹配模式的所有 key 并批量 DEL
func (s *CacheService) scanAndDel(ctx context.Context, pattern string) (int64, error) {
	var cursor uint64
	var deletedCount int64

	for {
		keys, nextCursor, err := s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return deletedCount, fmt.Errorf("redis scan failed: %w", err)
		}
		if len(keys) > 0 {
			deleted, err := s.rdb.Del(ctx, keys...).Result()
			if err != nil {
				return deletedCount, fmt.Errorf("redis del failed: %w", err)
			}
			deletedCount += deleted
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return deletedCount, nil
}
