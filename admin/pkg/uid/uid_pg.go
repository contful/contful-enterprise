// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

//go:build !dm

package uid

// GormDataType PostgreSQL native UUID 类型
func (UID) GormDataType() string { return "uuid" }

// GenUUID GORM default 值生成函数
func GenUUID() string { return "gen_random_uuid()" }
