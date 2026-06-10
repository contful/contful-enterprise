// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

//go:build dm

package uid

// GormDataType 达梦 CHAR(36) 类型
func (UID) GormDataType() string { return "char(36)" }

// GenUUID 达梦 SYS_GUID() → 标准 UUID 格式
func GenUUID() string { return "CONTFUL.GEN_UUID()" }
