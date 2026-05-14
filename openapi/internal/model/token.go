// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"github.com/google/uuid"
)

// TokenContext Token 验证通过后存入 Context 的信息
type TokenContext struct {
	TokenID   uuid.UUID `json:"token_id"`
	SiteID    uuid.UUID `json:"site_id"`
	Name      string    `json:"name"`
	ExpiresAt *int64    `json:"expires_at,omitempty"`
}

// Context keys
const (
	TokenContextKey = "api_token"
)
