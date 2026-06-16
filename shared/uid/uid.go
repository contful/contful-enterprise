// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package uid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
)

// UID 统一 UUID 类型，兼容 PostgreSQL（native UUID）和达梦（VARCHAR(36)）
// 运行时根据 database.SetDBType() 切换 GormDataType 行为
type UID uuid.UUID

// curDBType 运行时的数据库类型
var curDBType atomic.Value

// SetDBType 设置当前数据库类型（由 database.Open 调用）
func SetDBType(t string) { curDBType.Store(t) }

func dbType() string {
	if v := curDBType.Load(); v != nil { return v.(string) }
	return "postgres"
}

// GormDataType 返回 GORM 列类型
func (UID) GormDataType() string {
	if dbType() == "dm" { return "char(36)" }
	return "uuid"
}

// GenUUID GORM default 值
func GenUUID() string {
	if dbType() == "dm" { return "CONTFUL_ENT.GEN_UUID()" }
	return "gen_random_uuid()"
}

// New 生成新 UID
func New() UID { return UID(uuid.New()) }

// Parse 解析字符串为 UID
func Parse(s string) (UID, error) {
	u, err := uuid.Parse(s)
	return UID(u), err
}

var Nil = UID(uuid.Nil)

func (u UID) String() string   { return uuid.UUID(u).String() }
func (u UID) IsNil() bool       { return u == Nil }
func (u *UID) Scan(src interface{}) error { return (*uuid.UUID)(u).Scan(src) }
func (u UID) Value() (driver.Value, error) { return uuid.UUID(u).Value() }
func (u UID) MarshalJSON() ([]byte, error) { return json.Marshal(uuid.UUID(u).String()) }
func (u *UID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil { return err }
	parsed, err := uuid.Parse(s)
	if err != nil { return fmt.Errorf("uid: invalid UUID: %s", s) }
	*u = UID(parsed)
	return nil
}
