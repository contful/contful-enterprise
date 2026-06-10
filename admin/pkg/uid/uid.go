// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package uid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// UID 统一 UUID 类型，兼容 PostgreSQL（native UUID）和达梦（VARCHAR(36)）
// 两种模式通过 build tag 选择 GormDataType() 返回值：
//   - 默认（!dm）: 返回 "uuid"
//   - 加 -tags dm: 返回 "char(36)"  + Default 为 GenUUID()
type UID uuid.UUID

// New 生成新 UID
func New() UID { return UID(uuid.New()) }

// Parse 解析字符串为 UID
func Parse(s string) (UID, error) {
	u, err := uuid.Parse(s)
	return UID(u), err
}

// Nil 返回空 UID（用于判断）
var Nil = UID(uuid.Nil)

// String 实现 fmt.Stringer
func (u UID) String() string { return uuid.UUID(u).String() }

// IsNil 判断是否为空
func (u UID) IsNil() bool { return u == Nil }

// Scan 实现 sql.Scanner（从 DB 读到 Go）
func (u *UID) Scan(src interface{}) error {
	return (*uuid.UUID)(u).Scan(src)
}

// Value 实现 driver.Valuer（从 Go 写到 DB）
func (u UID) Value() (driver.Value, error) {
	return uuid.UUID(u).Value()
}

// MarshalJSON 实现 JSON 序列化
func (u UID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.UUID(u).String())
}

// UnmarshalJSON 实现 JSON 反序列化
func (u *UID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return fmt.Errorf("uid: invalid UUID: %s", s)
	}
	*u = UID(parsed)
	return nil
}
