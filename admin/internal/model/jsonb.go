// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB JSONB 类型
type JSONB map[string]interface{}

// Scan 实现 sql.Scanner 接口（兼容 []byte 和 string，PostgreSQL 不同驱动返回类型不同）
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported JSONB scan type")
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Map 转换为 map
func (j JSONB) Map() map[string]interface{} {
	return j
}

// Interface 返回接口
func (j JSONB) Interface() interface{} {
	return j
}

// JSONBSlice JSONB 数组类型
type JSONBSlice []map[string]interface{}

// Scan 实现 sql.Scanner 接口（兼容 []byte 和 string）
func (j *JSONBSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported JSONB scan type")
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONBSlice) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// JSONArray JSONB 数组类型（用于存储数组数据）
type JSONArray []interface{}

// Scan 实现 sql.Scanner 接口（兼容 []byte 和 string）
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported JSONB scan type")
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}
