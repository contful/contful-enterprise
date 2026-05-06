// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// CacheService Admin 缓存服务（用于清除 OpenAPI 缓存）
type CacheService struct {
	rdb *redis.Client
}

// NewCacheService 创建 CacheService
func NewCacheService(rdb *redis.Client) *CacheService {
	return &CacheService{rdb: rdb}
}

// CacheKeyPrefix 缓存键前缀（需与 OpenAPI 保持一致）
const AdminCacheKeyPrefix = "contful:content:"

// InvalidateSite 清除指定站点的所有内容缓存
func (s *CacheService) InvalidateSite(ctx context.Context, siteID string) (int64, error) {
	pattern := fmt.Sprintf("%s*:%s:*", AdminCacheKeyPrefix, siteID)

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

	log.Printf("[Cache] Invalidated %d keys for site %s", deletedCount, siteID)
	return deletedCount, nil
}
