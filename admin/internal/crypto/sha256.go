// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
)

// SHA256Hasher SHA-256 哈希器
type SHA256Hasher struct{}

// Sum 计算 SHA-256 哈希，返回 32 字节
func (h *SHA256Hasher) Sum(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HMAC 计算 HMAC-SHA256
func (h *SHA256Hasher) HMAC(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
