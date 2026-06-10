// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package audit

import (
	"context"

	"github.com/contful/contful/admin/pkg/uid"
)

// DataSigner 数据签名接口。
// 项目默认使用 HMAC-SHA256，用户可实现此接口替换为自有签名方法。
type DataSigner interface {
	Sign(entityType string, entityID uid.UID, payload string) (string, error)
	Verify(entityType string, entityID uid.UID, payload string, signature string) (bool, error)
	Algorithm() string
}

// signerCtxKey context key
type signerCtxKey struct{}

// WithSigner 将 DataSigner 注入 context
func WithSigner(ctx context.Context, s DataSigner) context.Context {
	return context.WithValue(ctx, signerCtxKey{}, s)
}

// GetSigner 从 context 取出 DataSigner（可能为 nil）
func GetSigner(ctx context.Context) DataSigner {
	if v := ctx.Value(signerCtxKey{}); v != nil {
		return v.(DataSigner)
	}
	return nil
}
